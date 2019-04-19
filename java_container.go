package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/si9ma/KillOJ-sandbox/model"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const (
	securityPolicy         = ``
	securityPolicyFile     = `security.policy`
	javaMemoryLimitMsg     = `Too small initial heap`               // todo fix hard code
	javaSecurityManagerMsg = `java.security.AccessControlException` // todo fix hard code
)

var javaCmd = cli.Command{
	Name:        "java",
	Usage:       "run java javaContainerlication",
	Description: `The java command use java security manager to run java javaContainerlication'`,

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
			Name:  "class",
			Usage: "the class name of java javaContainerlication",
		},
		cli.StringFlag{
			Name:  "expected",
			Value: "",
			Usage: "expected of test case",
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
		javaContainer := NewJavaContainer(ctx)
		defer javaContainer.handleResult()

		if javaContainer.err != nil {
			return nil // error from NewJavaContainer(ctx),return nil, handle error by self
		}

		// input/dir/expected is required
		if javaContainer.err = checkCmdStrArgsExist(ctx, []string{"input", "dir", "expected", "class"}); javaContainer.err != nil {
			return nil // argument error,return nil, handle error by self
		}

		time.AfterFunc(time.Duration(javaContainer.timeout)*time.Millisecond, func() {
			if javaContainer.cmd.Process != nil {
				APP_TIMEOUT_ERR = fmt.Errorf("Time Out")
				javaContainer.err = APP_TIMEOUT_ERR
				_ = syscall.Kill(-javaContainer.cmd.Process.Pid, syscall.SIGKILL)
			}
		})

		startTime := time.Now().UnixNano() / int64(time.Millisecond)
		if err := javaContainer.cmd.Run(); err != nil {
			APP_RUN_ERR = err
			javaContainer.err = APP_RUN_ERR
		}
		endTime := time.Now().UnixNano() / int64(time.Millisecond)
		javaContainer.timeCost = endTime - startTime

		return nil // return nil, handle error by self
	},
}

type JavaContainer struct {
	id         string
	err        error
	cmd        *exec.Cmd // compiler cmd
	class      string
	baseDir    string
	stdOut     bytes.Buffer
	stdErr     bytes.Buffer
	timeout    int64 // timeout in ms
	memory     int64
	memoryCost int64  // memory usage in KB
	timeCost   int64  // time usage in ms
	input      string // input of test case
	expected   string // expected of test case
}

func NewJavaContainer(ctx *cli.Context) *JavaContainer {
	javaContainer := &JavaContainer{
		id:       getGlbString(ctx, "id"),
		baseDir:  getString(ctx, "dir"),
		timeout:  ctx.Int64("timeout"),
		memory:   ctx.Int64("memory"),
		class:    getString(ctx, "class"),
		expected: getString(ctx, "expected"),
		input:    getString(ctx, "input"),
	}

	// write security.policy file
	if err := ioutil.WriteFile(securityPolicyFile, []byte(securityPolicy), 0644); err != nil {
		javaContainer.err = fmt.Errorf("newJavaContainer:%s", err.Error())
		return javaContainer
	}

	args := []string{
		"-Djava.security.manager",
		"-Djava.security.policy==" + securityPolicyFile,
		"-Xmx" + strconv.FormatInt(javaContainer.memory, 10) + "k", // limit memory
		javaContainer.class,
	}
	javaContainer.cmd = exec.Command("/usr/bin/java", args...)

	javaContainer.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	javaContainer.cmd.Dir = javaContainer.baseDir
	javaContainer.cmd.Stdout = &javaContainer.stdOut
	javaContainer.cmd.Stderr = &javaContainer.stdErr

	return javaContainer
}

func (j *JavaContainer) handleResult() {
	result := &model.RunResult{
		Result: model.Result{
			ID:         j.id,
			ResultType: model.RunResType,
			StdErr:     j.stdErr.String(),
		},
		Runtime:  j.timeCost,
		Input:    j.input,
		Expected: j.expected,
		Output:   j.stdOut.String(),
	}

	if j.cmd.ProcessState != nil {
		result.Memory = j.cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss
	}

	// handle error
	if j.err != nil {
		result.Message = j.err.Error()
		result.Status = model.FAIL

		switch j.err {
		case APP_TIMEOUT_ERR:
			result.Errno = model.RUN_TIMEOUT
		case APP_RUN_ERR:
			j.handleAppError(result)
		default:
			result.Errno = model.RUNNER_ERR
		}
	} else {
		if result.Expected == result.Output {
			result.Status = model.SUCCESS
			result.Message = "success"
		} else {
			result.Status = model.FAIL
			result.Errno = model.UNEXPECTED_RES_ERR
			result.Message = "output is unexpected"
		}
	}

	res, _ := json.Marshal(result)
	fmt.Println(string(res))

	// log result
	log.WithFields(log.Fields{
		"id":       j.id,
		"input":    j.input,
		"baseDir":  j.baseDir,
		"memory":   j.memory,
		"timeout":  j.timeout,
		"expected": j.expected,
		"class":    j.class,
		"result":   result,
	}).Info("run result")
}

func (j *JavaContainer) handleAppError(result *model.RunResult) {
	out := j.stdOut.String()
	err := j.stdErr.String()

	if strings.Contains(out, javaMemoryLimitMsg) {
		result.Errno = model.OUT_OF_MEMORY
		result.Message = "out of memory"
		return
	}

	if strings.Contains(err, javaSecurityManagerMsg) {
		result.Errno = model.JAVA_SECURITY_MANAGER_ERR
		result.Message = "security manager access denied"
		return
	}

	result.Errno = model.APP_ERR
}
