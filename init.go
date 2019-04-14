package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	libseccomp "github.com/seccomp/libseccomp-golang"
	"github.com/si9ma/KillOJ-sandbox/model"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var seccompMap = map[string][]string{
	"default": {
		"access", "arch_prctl", "brk", "clone", "close", "execve", "exit_group", "fcntl", "fstat",
		"futex", "getdents", "getpid", "getuid", "ioctl", "lstat", "mmap", "mprotect", "open", "openat",
		"pause", "pipe2", "poll", "read", "rt_sigaction", "rt_sigprocmask", "rt_sigreturn", "select",
		"setuid", "setxattr", "signalfd4", "stat", "statfs", "statfs64", "swapoff", "symlink",
		"sync_file_range", "sysfs", "tgkill", "timerfd_create", "uname", "ustat", "utime", "wait4", "waitid", "write",
	},
	"test": {
		"alarm",
	},
}

var RUNERR error

var initCmd = cli.Command{
	Name:        "init",
	Usage:       "init and run container",
	Description: `The init command use to init and run container, don't use this command from command line'`,

	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "input",
			Value: "",
			Usage: `input of test case`,
		},
		cli.StringFlag{
			Name:  "base-dir,dir",
			Value: "",
			Usage: "base directory of source code",
		},
		cli.StringFlag{
			Name:  "cmd",
			Usage: "the command name run in container",
		},
		cli.StringFlag{
			Name:  "expected",
			Value: "",
			Usage: "expected of test case",
		},
		cli.BoolFlag{
			Name:  "seccomp",
			Usage: "whether or not enable seccomp",
		},
		cli.IntFlag{
			Name:  "timeout",
			Value: 2000,
			Usage: "timeout limitation in milliseconds",
		},
		cli.IntFlag{
			Name:  "memory",
			Value: 256,
			Usage: "memory limitation in KB",
		},
	},
	Action: func(ctx *cli.Context) error {
		app := NewApp(ctx)
		defer app.HandleResult()

		// init namespace
		if app.err = initNamespace(app.dir); app.err != nil {
			return nil // return nil, handle error by self
		}

		// set rlimit
		// memory limit
		//memLimit := app.memory * 1024
		//if app.err = syscall.Setrlimit(syscall.RLIMIT_AS,&syscall.Rlimit{Max:memLimit,Cur:memLimit});app.err != nil {
		//	return nil // return nil, handle error by self
		//}
		//// max file size limit, use to disable create file
		//if app.err = syscall.Setrlimit(syscall.RLIMIT_FSIZE,&syscall.Rlimit{Max:0,Cur:0});app.err != nil {
		//	return nil // return nil, handle error by self
		//}

		// when seccomp is enabled
		if app.scmp {
			var scmpFilter *libseccomp.ScmpFilter
			if scmpFilter, app.err = enableSeccomp("test"); app.err != nil {
				return nil
			}
			if scmpFilter != nil {
				defer scmpFilter.Release()
			}
		}

		// time limit
		time.AfterFunc(time.Duration(app.timeout)*time.Millisecond, func() {
			if app.command.Process != nil {
				// use SIGUSR1 as time limit kill signal
				_ = syscall.Kill(-app.command.Process.Pid, syscall.SIGUSR1)
			}
		})

		// run app and calculate time
		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		if app.err = app.command.Run(); app.err != nil {
			RUNERR = fmt.Errorf("%s", app.err.Error())
			app.err = RUNERR
		}
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		app.timeCost = endTime - startTime

		return nil // return nil, handle error by self
	},
}

type App struct {
	id             string
	err            error  // run error
	input          string // input of test case
	dir            string // base dir
	expected       string // expected of test case
	timeout        int64  // timeout limit in ms
	memory         uint64 // memory limit in KB
	scmp           bool   // enable sccomp
	cmdStr         string // command of app
	command        *exec.Cmd
	stdOut         bytes.Buffer
	stdErr         bytes.Buffer
	memoryCost     int64 // memory usage in KB
	timeCost       int64 // time usage in ms
	rusage         *syscall.Rusage
	waitStatus     syscall.WaitStatus
	processStarted bool // if process is started
}

func NewApp(ctx *cli.Context) *App {
	app := &App{
		id:       ctx.GlobalString("id"),
		input:    ctx.String("input"),
		dir:      ctx.String("dir"),
		expected: ctx.String("expected"),
		timeout:  ctx.Int64("timeout"),
		memory:   ctx.Uint64("memory"),
		scmp:     ctx.Bool("seccomp"),
		cmdStr:   ctx.String("cmd"),
	}

	app.initCommand()

	return app
}

func (app *App) HandleResult() {
	result := model.RunResult{
		Result: model.Result{
			ID:         app.id,
			ResultType: model.RunResType,
		},
		Runtime:  app.timeCost,
		Input:    app.input,
		Output:   app.stdOut.String(),
		Expected: app.expected,
	}

	// get rusage and wait status info
	if app.command.ProcessState != nil {
		app.processStarted = true
		app.rusage = app.command.ProcessState.SysUsage().(*syscall.Rusage)
		app.waitStatus = app.command.ProcessState.Sys().(syscall.WaitStatus)
		result.Memory = app.rusage.Maxrss / 1024
	}

	// error is not nil or exit status is not 0
	if app.err != nil || app.processStarted && app.waitStatus.ExitStatus() != 0 {
		app.HandleError(&result)
	} else {
		// success
		if result.Output == result.Expected {
			result.Status = model.SUCCESS
		} else {
			result.Status = model.FAIL
			result.Errno = model.UNEXPECTED_RES_ERR
			result.Message = "output is unexpected"
		}
	}

	app.Log(result)
}

func (app *App) Log(result model.RunResult) {
	resultStr, _ := json.Marshal(result)
	fmt.Println(string(resultStr))

	// log result
	log.WithFields(log.Fields{
		"id":       app.id,
		"input":    app.input,
		"baseDir":  app.dir,
		"memory":   app.memory,
		"timeout":  app.timeout,
		"scmp":     app.scmp,
		"expected": app.expected,
		"cmdStr":   app.cmdStr,
		"result":   result,
	}).Info("app result")

}

func (app *App) HandleError(result *model.RunResult) {
	result.Status = model.FAIL

	// handle kill signal
	if app.processStarted && app.waitStatus.ExitStatus() != 0 {
		switch app.waitStatus.Signal() {
		case syscall.SIGSYS:
			result.Errno = model.BAD_SYSTEMCALL
			result.Message = "Bad System Call"
		case syscall.SIGUSR1:
			result.Errno = model.TIMEOUT
			result.Message = "Time Out"
		default:
			goto NotSigned
		}
		return
	}

NotSigned:

	// container error
	if app.err != nil && app.err != RUNERR {
		result.Errno = model.CONTAINER_ERR
		result.Message = app.err.Error()
		return
	}

	if app.err == RUNERR {
		result.Errno = model.APP_ERR
		result.Message = app.err.Error()
		return
	}
}

func (app *App) initCommand() {
	command := exec.Command(app.cmdStr)
	command.Stdin = strings.NewReader(app.input)
	command.Stdout = &app.stdOut
	command.Stderr = &app.stdErr
	command.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	PS1 := fmt.Sprintf("PS1=[%s] #", appName)
	PATH := fmt.Sprintf("PATH=PATH=/usr/sbin:/usr/bin:/sbin:/bin:/root/bin")
	command.Env = []string{PS1, PATH}

	app.command = command
}

func initNamespace(newRoot string) error {
	if err := pivotRoot(newRoot); err != nil {
		return fmt.Errorf("init namespace:%s", err.Error())
	}

	if err := syscall.Sethostname([]byte(appName)); err != nil {
		return fmt.Errorf("init namespace:%s", err.Error())
	}

	return nil
}

func pivotRoot(newRoot string) error {
	putOld := filepath.Join(newRoot, "/.pivot_root")

	// bind mount new_root to itself - this is a slight hack needed to satisfy requirement (2)
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	// create put_old directory
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	// call pivotRoot
	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	// Note that this also applies to the calling process: pivotRoot() may
	// or may not affect its current working directory.  It is therefore
	// recommended to call chdir("/") immediately after pivotRoot().
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	// umount put_old, which now lives at /.pivot_root
	putOld = "/.pivot_root"
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	// remove put_old
	if err := os.RemoveAll(putOld); err != nil {
		return fmt.Errorf("pivot_root:%s", err.Error())
	}

	return nil
}

func enableSeccomp(config string) (*libseccomp.ScmpFilter, error) {
	var syscalls []string
	if val, ok := seccompMap[config]; !ok {
		return nil, fmt.Errorf("seccomp:config %s is not exist", config)
	} else {
		syscalls = val
	}

	// new filter
	//filter,err := libseccomp.NewFilter(libseccomp.ActErrno.SetReturnCode(int16(syscall.EPERM)))
	filter, err := libseccomp.NewFilter(libseccomp.ActAllow)
	if err != nil {
		return nil, fmt.Errorf("seccomp:%s", err.Error())
	}

	// add arch
	archs := []string{"x86", "x86_64", "x32"}
	for _, arch := range archs {
		scmpArch, err := libseccomp.GetArchFromString(arch)
		if err != nil {
			return nil, fmt.Errorf("seccomp:%s", err.Error())
		}

		if err := filter.AddArch(scmpArch); err != nil {
			return nil, fmt.Errorf("seccomp:%s", err.Error())
		}
	}

	// set No New Privileges bit
	if err := filter.SetNoNewPrivsBit(true); err != nil {
		return nil, fmt.Errorf("seccomp:%s", err.Error())
	}

	// add whitelist rule
	for _, item := range syscalls {
		syscallID, err := libseccomp.GetSyscallFromName(item)
		if err != nil {
			return nil, fmt.Errorf("seccomp:%s", err)
		}
		if err := filter.AddRule(syscallID, libseccomp.ActKill); err != nil {
			return nil, fmt.Errorf("seccomp:%s", err)
		}
	}

	// load
	if err := filter.Load(); err != nil {
		return nil, fmt.Errorf("seccomp:%s", err.Error())
	}

	return filter, nil
}