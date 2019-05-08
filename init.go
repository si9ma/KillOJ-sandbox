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

	"github.com/si9ma/KillOJ-common/tip"

	"github.com/si9ma/KillOJ-common/judge"

	libseccomp "github.com/seccomp/libseccomp-golang"
	"github.com/urfave/cli"
)

var APP_RUN_ERR error
var APP_TIMEOUT_ERR error

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
			Value: 6000,
			Usage: "timeout limitation in milliseconds",
		},
		cli.IntFlag{
			Name:  "memory",
			Value: 256 * 1024,
			Usage: "memory limitation in KB",
		},
	},
	Action: func(ctx *cli.Context) error {
		app := NewApp(ctx)
		defer app.handleResult()

		// init namespace
		if app.err = initNamespace(app.dir); app.err != nil {
			return nil // return nil, handle error by self
		}

		// set rlimit
		// memory limit
		//memLimit := app.memory * 1024
		//if app.err = syscall.Setrlimit(syscall.RLIMIT_AS, &syscall.Rlimit{Max: memLimit, Cur: memLimit}); app.err != nil {
		//	return nil // return nil, handle error by self
		//}

		// when seccomp is enabled
		if app.scmp {
			var scmpFilter *libseccomp.ScmpFilter
			if scmpFilter, app.err = enableSeccomp("blacklist-default", app.cmdStr); app.err != nil {
				return nil
			}
			scmpFilter.Release()
		}

		// time out
		time.AfterFunc(time.Duration(app.timeout)*time.Millisecond, func() {
			if app.command.Process != nil {
				APP_TIMEOUT_ERR = fmt.Errorf("Time Out")
				app.err = APP_TIMEOUT_ERR
				_ = syscall.Kill(-app.command.Process.Pid, syscall.SIGKILL)
			}
		})

		// run app and calculate time
		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		if app.err = app.command.Run(); app.err != nil {
			APP_RUN_ERR = fmt.Errorf("%s", app.err.Error())
			app.err = APP_RUN_ERR
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
	memLimit       int64  // memory limit in KB
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
		id:       getGlbString(ctx, "id"),
		input:    getString(ctx, "input"),
		dir:      getString(ctx, "dir"),
		expected: getString(ctx, "expected"),
		timeout:  ctx.Int64("timeout"),
		memLimit: ctx.Int64("memory"),
		scmp:     ctx.Bool("seccomp"),
		cmdStr:   getString(ctx, "cmd"),
	}

	app.initCommand()

	return app
}

func (app *App) handleResult() {
	result := judge.InnerResult{
		ID:         app.id,
		ResultType: judge.RunResType,
		StdErr:     app.stdErr.String(),
		Runtime:    app.timeCost,
		Input:      app.input,
		Output:     app.stdOut.String(),
		Expected:   app.expected,
		TimeLimit:  app.timeout,
		MemLimit:   app.memLimit,
	}

	// get rusage and wait status info
	if app.command.ProcessState != nil {
		app.processStarted = true
		app.rusage = app.command.ProcessState.SysUsage().(*syscall.Rusage)
		app.waitStatus = app.command.ProcessState.Sys().(syscall.WaitStatus)
		result.Memory = app.rusage.Maxrss
	}

	// error is not nil or exit status is not 0
	if app.err != nil || app.processStarted && app.waitStatus.ExitStatus() != 0 {
		app.handleError(&result)
	} else {
		// success
		if result.Output == result.Expected {
			result.Status = judge.SUCCESS
			result.Message = tip.RunSuccessTip.String()
		} else {
			result.Status = judge.FAIL
			result.Errno = judge.WRONG_ANSWER_ERR
			result.Message = judge.GetInnerErrorMsgByErrNo(judge.WRONG_ANSWER_ERR)
		}
	}

	resultStr, _ := json.Marshal(result)
	fmt.Println(string(resultStr))
}

func (app *App) handleError(result *judge.InnerResult) {
	result.Status = judge.FAIL

	// handle kill signal
	if app.processStarted && app.waitStatus.ExitStatus() != 0 {
		switch app.waitStatus.Signal() {
		case syscall.SIGSYS:
			result.Errno = judge.BAD_SYSTEMCALL_ERR
			result.Message = judge.GetInnerErrorMsgByErrNo(judge.BAD_SYSTEMCALL_ERR)
		case syscall.SIGKILL: // kill by oomkiller
			result.Errno = judge.OUT_OF_MEMORY_ERR
			result.Message = judge.GetInnerErrorMsgByErrNo(judge.OUT_OF_MEMORY_ERR)
		default:
			goto NotSigned
		}
		return
	}

NotSigned:

	result.Message = app.err.Error()
	switch app.err {
	case APP_TIMEOUT_ERR:
		result.Errno = judge.RUN_TIMEOUT_ERR
	case APP_RUN_ERR:
		result.Errno = judge.APP_ERR
	default:
		result.Errno = judge.CONTAINER_ERR
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
