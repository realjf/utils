package utils

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type Command struct {
	uid_ui32    uint32
	gid_ui32    uint32
	uid         int
	gid         int
	user        *user.User
	debug       bool
	stdoutbuf   bytes.Buffer
	stderrbuf   bytes.Buffer
	wg          sync.WaitGroup
	pid         int
	cmd         *exec.Cmd
	timeout     time.Duration
	workDir     string
	ctx         context.Context
	cancel      context.CancelFunc
	procAttr    *syscall.SysProcAttr
	credential  *syscall.Credential
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	stdin       io.WriteCloser
	stdinChan   chan string
	running     bool
	lock        sync.Mutex
	env         []string
	noSetGroups bool
}

func NewCmd() *Command {
	return &Command{}
}

func NewCommand(uid int, gid int, user *user.User) *Command {
	return &Command{
		uid_ui32:  uint32(uid),
		gid_ui32:  uint32(gid),
		uid:       uid,
		gid:       gid,
		user:      user,
		debug:     false,
		stdoutbuf: bytes.Buffer{},
		stderrbuf: bytes.Buffer{},
		wg:        sync.WaitGroup{},
		pid:       0,
		cmd:       &exec.Cmd{},
		timeout:   0,
		workDir:   "",
		// procAttr: &syscall.SysProcAttr{
		// Cloneflags:                 syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
		// GidMappingsEnableSetgroups: true,
		// Setpgid: true,
		// UidMappings: []syscall.SysProcIDMap{
		// 	{
		// 		ContainerID: 0,
		// 		HostID:      0,
		// 		Size:        1,
		// 	},
		// },
		// GidMappings: []syscall.SysProcIDMap{
		// 	{
		// 		ContainerID: 0,
		// 		HostID:      0,
		// 		Size:        1,
		// 	},
		// },
		// 	Pgid:       0,
		// 	Credential: &syscall.Credential{},
		// },
		procAttr:    nil,
		credential:  nil,
		stdout:      nil,
		stderr:      nil,
		stdin:       nil,
		stdinChan:   make(chan string),
		running:     false,
		lock:        sync.Mutex{},
		env:         nil,
		noSetGroups: false,
	}
}

func (c *Command) SetSysProcAttr(procAttr syscall.SysProcAttr) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.procAttr = &procAttr
	return c
}

func (c *Command) SetSysCredential(credential syscall.Credential) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.credential = &credential
	return c
}

func (c *Command) SetTimeout(t time.Duration) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.timeout = t
	return c
}

func (c *Command) SetWorkDir(wd string) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.workDir = wd
	return c
}

func (c *Command) SetDebug(debug bool) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.debug = debug
	return c
}

func (c *Command) SetUser(u *user.User) {
	c.lock.Lock()
	defer c.lock.Unlock()
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
	c.lock.Lock()
	defer c.lock.Unlock()
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
	c.lock.Lock()
	defer c.lock.Unlock()
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
	_, err = c.Command(cmdl, args...)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

func (c *Command) GetPid() int {
	return c.pid
}

func (c *Command) Resume() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// resume subprocess
	return c.cmd.Process.Signal(syscall.SIGCONT)
}

func (c *Command) checkProcStateIsRunning() {
	for {
		if c.running {
			run, err := CheckProcStateIsStopped(c.pid)
			if err != nil {
				if c.debug {
					log.Error(err)
				}

			}
			if !run {
				err = c.Resume()
				if err != nil {
					if errors.Is(err, os.ErrProcessDone) {
						break
					}
					if c.debug {
						log.Error(err)
					}
				}
				c.running = true
			} else {
				break
			}
		} else {
			break
		}
		time.Sleep(3 * time.Second)
	}
}

func (c *Command) Run() (output []byte, err error) {
	if c.cmd.Process == nil {
		if c.debug {
			log.Error("subprocess already exited")
		}
		return nil, errors.New("subprocess already exited")
	}
	err = c.Resume()
	if err != nil {
		if !errors.Is(err, os.ErrProcessDone) {
			return nil, err
		}
	}
	c.wg.Wait()

	if err = c.cmd.Wait(); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			if c.debug {
				log.Infof("Process done: %s", err.Error())
			}
			return c.stdoutbuf.Bytes(), nil
		}
		if exiterr, ok := err.(*exec.ExitError); ok {
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				if c.debug {
					log.Infof("Exit Status: %d", status.ExitStatus())
				}

				return c.stdoutbuf.Bytes(), err
			}
		}
		if c.debug {
			log.Error(err.Error())
		}
		return c.stdoutbuf.Bytes(), err
	}

	return c.GetOutput()
}

func (c *Command) Close() {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.stdout != nil {
		c.stdout.Close()
	}
	if c.stderr != nil {
		c.stderr.Close()
	}
	if c.stdin != nil {
		c.stdin.Close()
	}
	if c.cancel != nil {
		c.cancel()
	}
}

func (c *Command) SetEnv(envs []string) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.env = envs
	return c
}

func (c *Command) SetNoSetGroups(noSetGroups bool) *Command {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.noSetGroups = noSetGroups
	return c
}

func (c *Command) Command(cmdl string, args ...string) (pid int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.debug {
		log.Infof("run command under the uid[%d] gid[%d]", c.uid_ui32, c.gid_ui32)
	}

	if c.timeout > 0 {
		c.ctx, c.cancel = context.WithTimeout(context.Background(), c.timeout)
		c.cmd = exec.CommandContext(c.ctx, cmdl, args...)
	} else {
		c.cmd = exec.Command(cmdl, args...)
	}

	if c.procAttr != nil {
		c.cmd.SysProcAttr = c.procAttr
	}

	if c.credential != nil {
		if c.cmd.SysProcAttr == nil {
			c.cmd.SysProcAttr = &syscall.SysProcAttr{}
		}
		c.cmd.SysProcAttr.Credential = c.credential
	}

	if c.uid_ui32 > 0 && c.gid_ui32 > 0 {
		if c.credential == nil {
			if c.cmd.SysProcAttr == nil {
				c.cmd.SysProcAttr = &syscall.SysProcAttr{}
			}
			c.cmd.SysProcAttr.Credential = &syscall.Credential{
				Uid:         c.uid_ui32,
				Gid:         c.gid_ui32,
				NoSetGroups: false, //
			}
		}
	}

	if c.noSetGroups {
		if c.cmd.SysProcAttr != nil && c.cmd.SysProcAttr.Credential != nil {
			c.cmd.SysProcAttr.Credential.NoSetGroups = c.noSetGroups
		}
	}

	if c.user != nil {
		c.cmd.Env = append(os.Environ(), "USER="+c.user.Username, "HOME="+c.user.HomeDir)
	} else {
		c.cmd.Env = os.Environ()
	}

	if c.env == nil {
		c.cmd.Env = c.env
	}

	if c.workDir != "" {
		c.cmd.Dir = c.workDir
	} else {
		curDir, err := os.Getwd()
		if err != nil {
			return pid, err
		}
		c.cmd.Dir = curDir
	}

	syscall.Setgid(c.gid)
	syscall.Setuid(c.uid)
	syscall.Setreuid(-1, c.uid)
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}
	stdoutReader := bufio.NewReader(c.stdout)

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}
	stderrReader := bufio.NewReader(c.stderr)

	c.stdin, err = c.cmd.StdinPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}

	if err = c.cmd.Start(); err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}

	c.pid = c.cmd.Process.Pid
	err = c.Pause()
	if err != nil {
		return pid, err
	}
	c.running = false
	c.wg.Add(1)
	go c.handleReader(stdoutReader, 1)
	c.wg.Add(1)
	go c.handleReader(stderrReader, 2)
	go c.checkProcStateIsRunning()
	return c.pid, nil
}

func (c *Command) Pause() error {
	// pause subprocess
	return c.cmd.Process.Signal(syscall.SIGTSTP)
}

func (c *Command) GetOutput() ([]byte, error) {
	if c.stderrbuf.Bytes() != nil {
		return nil, fmt.Errorf("%s", c.stderrbuf.Bytes())
	}
	return c.stdoutbuf.Bytes(), nil
}

func (c *Command) GetStderrOutput() ([]byte, error) {
	return c.stderrbuf.Bytes(), nil
}

func (c *Command) NeedInput(text string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	stdinWriter := bufio.NewWriter(c.stdin)
	reader := bufio.NewReader(os.Stdin)
	os.Stdout.WriteString(text)

	text, err := reader.ReadString('\n')
	if err != nil {
		log.Error(err)
	}
	_, err = io.WriteString(stdinWriter, text)
	if err != nil {
		log.Error(err)
	}
}

func (c *Command) handleReader(reader *bufio.Reader, stdio int) {
	defer func() {
		c.wg.Done()
	}()
	for {
		str, err := reader.ReadString('\n')
		if stdio == 1 {
			c.stdoutbuf.WriteString(str)
		} else if stdio == 2 {
			c.stderrbuf.WriteString(str)
		}
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				if c.debug {
					log.Info(err)
				}
				return
			} else if errors.Is(err, io.EOF) {
				if c.debug {
					log.Info("Read EOF")
				}
				// return
				continue
			} else if errors.Is(err, os.ErrClosed) {
				if c.debug {
					log.Info(err)
				}
				return
			} else if errors.Is(err, os.ErrDeadlineExceeded) {
				if c.debug {
					log.Info(err)
				}
				return
			} else if errors.Is(err, io.ErrUnexpectedEOF) {
				if c.debug {
					log.Info(err)
				}
				return
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
