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

	"github.com/si9ma/KillOJ-common/judge"

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
	id          string
	err         error
	cmd         *exec.Cmd // compiler cmd
	class       string
	baseDir     string
	stdOut      bytes.Buffer
	stdErr      bytes.Buffer
	timeout     int64 // timeout in ms
	memoryLimit int64
	memoryCost  int64  // memory usage in KB
	timeCost    int64  // time usage in ms
	input       string // input of test case
	expected    string // expected of test case
}

func NewJavaContainer(ctx *cli.Context) *JavaContainer {
	javaContainer := &JavaContainer{
		id:          getGlbString(ctx, "id"),
		baseDir:     getString(ctx, "dir"),
		timeout:     ctx.Int64("timeout"),
		memoryLimit: ctx.Int64("memory"),
		class:       getString(ctx, "class"),
		expected:    getString(ctx, "expected"),
		input:       getString(ctx, "input"),
	}

	// write security.policy file
	if err := ioutil.WriteFile(securityPolicyFile, []byte(securityPolicy), 0644); err != nil {
		javaContainer.err = fmt.Errorf("newJavaContainer:%s", err.Error())
		return javaContainer
	}

	args := []string{
		"-Djava.security.manager",
		"-Djava.security.policy==" + securityPolicyFile,
		"-Xmx" + strconv.FormatInt(javaContainer.memoryLimit, 10) + "k", // limit memory
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
	result := &judge.InnerResult{
		ID:         j.id,
		ResultType: judge.RunResType,
		StdErr:     j.stdErr.String(),
		Runtime:    j.timeCost,
		Input:      j.input,
		Expected:   j.expected,
		Output:     j.stdOut.String(),
		TimeLimit:  j.timeout,
		MemLimit:   j.memoryLimit,
	}

	if j.cmd.ProcessState != nil {
		result.Memory = j.cmd.ProcessState.SysUsage().(*syscall.Rusage).Maxrss
	}

	// handle error
	if j.err != nil {
		result.Message = j.err.Error()
		result.Status = judge.FAIL

		switch j.err {
		case APP_TIMEOUT_ERR:
			result.Errno = judge.RUN_TIMEOUT_ERR
		case APP_RUN_ERR:
			j.handleAppError(result)
		default:
			result.Errno = judge.RUNNER_ERR
		}
	} else {
		if result.Expected == result.Output {
			result.Status = judge.SUCCESS
			result.Message = judge.GetRunSuccessMsg()
		} else {
			result.Status = judge.FAIL
			result.Errno = judge.WRONG_ANSWER_ERR
			result.Message = judge.GetInnerErrorMsgByErrNo(judge.WRONG_ANSWER_ERR)
		}
	}

	res, _ := json.Marshal(result)
	fmt.Println(string(res))

	// log result
	log.WithFields(log.Fields{
		"id":       j.id,
		"input":    j.input,
		"baseDir":  j.baseDir,
		"memLimit": j.memoryLimit,
		"timeout":  j.timeout,
		"expected": j.expected,
		"class":    j.class,
		"result":   result,
	}).Info("run result")
}

func (j *JavaContainer) handleAppError(result *judge.InnerResult) {
	out := j.stdOut.String()
	err := j.stdErr.String()

	if strings.Contains(out, javaMemoryLimitMsg) {
		result.Errno = judge.OUT_OF_MEMORY_ERR
		result.Message = judge.GetInnerErrorMsgByErrNo(judge.OUT_OF_MEMORY_ERR)
		return
	}

	if strings.Contains(err, javaSecurityManagerMsg) {
		result.Errno = judge.JAVA_SECURITY_MANAGER_ERR
		result.Message = judge.GetInnerErrorMsgByErrNo(judge.JAVA_SECURITY_MANAGER_ERR)
		return
	}

	result.Errno = judge.APP_ERR
}
