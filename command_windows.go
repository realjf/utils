package utils

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/user"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"
)

type Cmd struct {
	user      *UserAccount
	debug     bool
	stdoutbuf *IOReadCloser
	stderrbuf *IOReadCloser
	wg        sync.WaitGroup
	pid       int
	cmd       *exec.Cmd
	timeout   time.Duration
	workDir   string
	ctx       context.Context
	cancel    context.CancelFunc
	procAttr  *syscall.SysProcAttr
	// credential  *syscall.Credential
	stdout      io.ReadCloser
	stderr      io.ReadCloser
	stdin       io.WriteCloser
	stdinChan   chan string
	running     bool
	lock        sync.RWMutex
	env         []string
	noSetGroups bool
	stop        chan bool
	kernel32DLL *syscall.LazyDLL
}

func NewCmd() *Cmd {
	c := &Cmd{
		user:      NewUserAccount(),
		debug:     false,
		stdoutbuf: nil,
		stderrbuf: nil,
		wg:        sync.WaitGroup{},
		pid:       0,
		cmd:       &exec.Cmd{},
		timeout:   0,
		workDir:   "",
		procAttr:  nil,
		// credential:  nil,
		stdout:      nil,
		stderr:      nil,
		stdin:       nil,
		stdinChan:   make(chan string),
		running:     false,
		lock:        sync.RWMutex{},
		env:         nil,
		noSetGroups: false,
		stop:        make(chan bool),
		kernel32DLL: syscall.NewLazyDLL("Kernel32.dll"),
	}
	runtime.SetFinalizer(c, (*Cmd).Close)
	return c
}

func NewCommand() *Cmd {
	c := &Cmd{
		user:      NewUserAccount(),
		debug:     false,
		stdoutbuf: nil,
		stderrbuf: nil,
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
		procAttr: nil,
		// credential:  nil,
		stdout:      nil,
		stderr:      nil,
		stdin:       nil,
		stdinChan:   make(chan string),
		running:     false,
		lock:        sync.RWMutex{},
		env:         nil,
		noSetGroups: false,
		stop:        make(chan bool),
		kernel32DLL: syscall.NewLazyDLL("Kernel32.dll"),
	}
	runtime.SetFinalizer(c, (*Cmd).Close)
	return c
}

func (c *Cmd) SetSysProcAttr(procAttr syscall.SysProcAttr) *Cmd {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.procAttr = &procAttr
	return c
}

// func (c *Cmd) SetSysCredential(credential syscall.Credential) *Cmd {
// 	c.lock.Lock()
// 	defer c.lock.Unlock()
// 	c.credential = &credential
// 	return c
// }

func (c *Cmd) SetTimeout(t time.Duration) *Cmd {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.timeout = t
	return c
}

func (c *Cmd) SetWorkDir(wd string) *Cmd {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.workDir = wd
	return c
}

func (c *Cmd) SetDebug(debug bool) *Cmd {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.debug = debug
	return c
}

func (c *Cmd) Mkdir(path string, perm os.FileMode) (output []byte, err error) {
	args := []string{"/c", "mkdir", "-p", path, "-m", "=" + perm.String()}
	output, err = c.RunCommand("cmd", args...)
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return
	}
	return
}

func (c *Cmd) RunCommand(cmdl string, args ...string) (output []byte, err error) {
	_, err = c.Command(cmdl, args...)
	if err != nil {
		return nil, err
	}
	return c.Run()
}

func (c *Cmd) GetPid() int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.pid
}

func (c *Cmd) Resume() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	// resume subprocess
	return c.cmd.Process.Signal(syscall.SIGCONT)
}

func (c *Cmd) checkProcStateIsRunning() {
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

func (c *Cmd) GetUser() *UserAccount {
	return c.user
}

func (c *Cmd) SetUser(u *user.User) {
	c.user.SetUser(u)
}

func (c *Cmd) SetUsername(username string) error {
	User, err := user.Lookup(username)
	if err != nil {
		return err
	}
	c.user.SetUser(User)
	return nil
}

func (c *Cmd) Run() (output []byte, err error) {
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

func (c *Cmd) Close() error {
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
	if c.lazyDLL != nil {
		return c.lazyDLL.Release()
	}
	return nil
}

func (c *Cmd) SetEnv(envs []string) *Cmd {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.env = envs
	return c
}

// func (c *Cmd) SetNoSetGroups(noSetGroups bool) *Cmd {
// 	c.lock.Lock()
// 	defer c.lock.Unlock()
// 	c.noSetGroups = noSetGroups
// 	return c
// }

func (c *Cmd) Command(cmdl string, args ...string) (pid int, err error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.debug {
		log.Infof("run command under the uid[%d] gid[%d]", c.GetUser().GetUid_ui32(), c.GetUser().GetGid_ui32())
	}

	arglist := []string{"/c", cmdl}
	arglist = append(arglist, args...)
	if c.timeout > 0 {
		c.ctx, c.cancel = context.WithTimeout(context.Background(), c.timeout)
		c.cmd = exec.CommandContext(c.ctx, "cmd", arglist...)
	} else {
		c.cmd = exec.Command("cmd", arglist...)
	}

	if c.procAttr != nil {
		c.cmd.SysProcAttr = c.procAttr
	}

	// if c.credential != nil {
	// 	if c.cmd.SysProcAttr == nil {
	// 		c.cmd.SysProcAttr = &syscall.SysProcAttr{}
	// 	}
	// 	// c.cmd.SysProcAttr.Credential = c.credential
	// }

	// if c.GetUser().GetUid_ui32() > 0 && c.GetUser().GetGid_ui32() > 0 {
	// 	if c.credential == nil {
	// 		if c.cmd.SysProcAttr == nil {
	// 			c.cmd.SysProcAttr = &syscall.SysProcAttr{}
	// 		}
	// 		c.cmd.SysProcAttr.Credential = &syscall.Credential{
	// 			Uid:         c.GetUser().GetUid_ui32(),
	// 			Gid:         c.GetUser().GetGid_ui32(),
	// 			NoSetGroups: false, //
	// 		}
	// 	}
	// }

	// if c.noSetGroups {
	// 	if c.cmd.SysProcAttr != nil && c.cmd.SysProcAttr.Credential != nil {
	// 		c.cmd.SysProcAttr.Credential.NoSetGroups = c.noSetGroups
	// 	}
	// }

	if c.GetUser().GetUser() != nil {
		c.cmd.Env = append(os.Environ(), "USER="+c.GetUser().GetUser().Username, "HOME="+c.GetUser().GetUser().HomeDir)
	} else {
		c.cmd.Env = os.Environ()
	}

	if c.env != nil {
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

	// syscall.Setgid(c.gid)
	// syscall.Setuid(c.uid)
	// syscall.Setreuid(-1, c.uid)
	c.stdout, err = c.cmd.StdoutPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}
	c.stdoutbuf = NewReader(c.stdout)

	c.stderr, err = c.cmd.StderrPipe()
	if err != nil {
		if c.debug {
			log.Error(err.Error())
		}
		return pid, err
	}
	c.stderrbuf = NewReader(c.stderr)

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
	go c.handleReader(c.stdout, STDOUT)
	c.wg.Add(1)
	go c.handleReader(c.stderr, STDERR)
	go c.checkProcStateIsRunning()
	return c.pid, nil
}

func (c *Cmd) Pause() error {
	// pause subprocess
	return c.cmd.Process.Signal(syscall.SIGTSTP)
}

func (c *Cmd) GetOutput() ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.stdoutbuf.Bytes(), nil
}

func (c *Cmd) GetStderrOutput() ([]byte, error) {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.stderrbuf.Bytes(), nil
}

func (c *Cmd) NeedInput(text string) {
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

func (c *Cmd) handleReader(std io.ReadCloser, stdio int) {
	defer func() {
		c.wg.Done()
	}()

	for {
		var err error
		var str string
		if stdio == STDOUT {
			str, err = c.stdoutbuf.Read()
		} else if stdio == STDERR {
			str, err = c.stderrbuf.Read()
		}
		if err != nil {
			if errors.Is(err, io.ErrClosedPipe) {
				if c.debug {
					log.Info(err)
				}
				return
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
			} else if errors.Is(err, io.EOF) {
				if c.debug {
					log.Info("Read EOF")
				}
				return
			} else {
				if c.debug {
					log.Error(err)
				}
				break
			}
		}

		if c.debug {
			log.Infof("%s", str)
		}
	}
}

func (c *Cmd) CheckRoot() error {
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
