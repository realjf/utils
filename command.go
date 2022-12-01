package utils

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Command struct {
	uid_ui32 uint32
	gid_ui32 uint32
	uid      int
	gid      int
	user     *user.User
	debug    bool
	stdout   bytes.Buffer
	stderr   bytes.Buffer
	wg       sync.WaitGroup
	pid      int
	cmd      *exec.Cmd
}

func NewCmd() *Command {
	return &Command{}
}

func NewCommand(uid int, gid int, user *user.User) *Command {
	return &Command{
		uid_ui32: uint32(uid),
		gid_ui32: uint32(gid),
		uid:      uid,
		gid:      gid,
		user:     user,
		debug:    false,
		stdout:   bytes.Buffer{},
		stderr:   bytes.Buffer{},
		wg:       sync.WaitGroup{},
	}
}

func (c *Command) SetDebug(debug bool) *Command {
	c.debug = debug
	return c
}

func (c *Command) SetUser(u *user.User) {
	c.user = u
	c.uid, _ = strconv.Atoi(u.Uid)
	c.gid, _ = strconv.Atoi(u.Gid)
	u64, _ := strconv.ParseUint(u.Uid, 10, 32)
	g64, _ := strconv.ParseUint(u.Gid, 10, 32)
	c.uid_ui32 = uint32(u64)
	c.gid_ui32 = uint32(g64)
}

func (c *Command) SetUsername(username string) error {
	User, err := user.Lookup(username)
	if err != nil {
		return err
	}
	c.SetUser(User)
	return nil
}

func (c *Command) Mkdir(path string, perm os.FileMode) (output []byte, err error) {
	args := []string{"-p", path, "-m", "=" + perm.String()}
	output, err = c.RunCommand("mkdir", args...)
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return
	}
	return
}

func (c *Command) Lsof(path string) (output []byte, err error) {
	args := []string{path}
	output, err = c.RunCommand("lsof", args...)
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return
	}
	return
}

func (c *Command) GetUser() *user.User {
	return c.user
}

func (c *Command) SetUid(u uint64) {
	c.uid = int(u)
	c.uid_ui32 = uint32(u)
}

func (c *Command) GetUid() int {
	return c.uid
}

func (c *Command) GetUid_ui32() uint32 {
	return c.uid_ui32
}

func (c *Command) SetGid(g uint64) {
	c.gid = int(g)
	c.gid_ui32 = uint32(g)
}

func (c *Command) GetGid() int {
	return c.gid
}

func (c *Command) GetGid_ui32() uint32 {
	return c.gid_ui32
}

func (c *Command) RunCommand(cmdl string, args ...string) (output []byte, err error) {
	err = c.Command(cmdl, args...)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

func (c *Command) GetPid() int {
	return c.pid
}

func (c *Command) Run() (output []byte, err error) {

	if err = c.cmd.Wait(); err != nil {
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if c.debug {
					log.Infof("Exit Status: %d", status.ExitStatus())
				}

				return c.stdout.Bytes(), err
			}
		}
		if c.debug {
			log.Error(err.Error())
		}
		return c.stdout.Bytes(), err
	}

	return c.GetOutput()
}

func (c *Command) Command(cmdl string, args ...string) (err error) {
	log.Infof("run command under the uid[%d] gid[%d]", c.uid_ui32, c.gid_ui32)
	c.cmd = exec.Command(cmdl, args...)
	c.cmd.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags:                 syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		GidMappingsEnableSetgroups: true,
		Setpgid:                    true,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      0,
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      0,
				Size:        1,
			},
		},
		Pgid: 0,
	}
	c.cmd.SysProcAttr.Credential = &syscall.Credential{
		Uid: c.uid_ui32,
		Gid: c.gid_ui32,
	}
	if c.user != nil {
		c.cmd.Env = append(os.Environ(), "USER="+c.user.Username, "HOME="+c.user.HomeDir)
	} else {
		c.cmd.Env = os.Environ()
	}

	syscall.Setgid(c.gid)
	syscall.Setuid(c.uid)
	syscall.Setreuid(-1, c.uid)
	stdout, err := c.cmd.StdoutPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return err
	}
	defer stdout.Close()
	stdoutReader := bufio.NewReader(stdout)

	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return err
	}
	defer stderr.Close()
	stderrReader := bufio.NewReader(stderr)

	if err = c.cmd.Start(); err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return err
	}

	c.pid = c.cmd.Process.Pid
	go c.handleReader(stdoutReader, 1)
	c.wg.Add(1)
	go c.handleReader(stderrReader, 2)
	c.wg.Add(1)
	c.wg.Wait()
	return nil
}

func (c *Command) GetOutput() ([]byte, error) {
	return c.stdout.Bytes(), nil
}

func (c *Command) GetStderrOutput() ([]byte, error) {
	return c.stderr.Bytes(), nil
}

func (c *Command) GetError() ([]byte, error) {
	return c.stderr.Bytes(), nil
}

func (c *Command) handleReader(reader *bufio.Reader, stdio int) {
	defer func() {
		c.wg.Done()
	}()
	for {
		str, err := reader.ReadString('\n')
		if stdio == 1 {
			c.stdout.WriteString(str)
		} else if stdio == 2 {
			c.stderr.WriteString(str)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				if c.debug {
					log.Info("Read EOF")
				}
				break
			}
			log.Error(err)
			break
		}

		if c.debug {
			log.Infof("%s", str)
		}
	}
}

func (c *Command) CheckRoot() error {
	user, err := user.Current()
	if err != nil {
		return err
	}
	uid, err := strconv.Atoi(user.Uid)
	if err != nil {
		return err
	}
	if uid != 0 {
		return errors.New("you must run as root")
	}
	return nil
}
