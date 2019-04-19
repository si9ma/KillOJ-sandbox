package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	uuid "github.com/satori/go.uuid"
	"github.com/si9ma/KillOJ-sandbox/cgroups"
	"github.com/si9ma/KillOJ-sandbox/model"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

const MemoryUsedByContainer = 1550 // 1550KB memory used by container

var runCmd = cli.Command{
	Name:        "run",
	Usage:       "run app in container",
	Description: `The run command use to run code in container`,

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
			Name:  "expected",
			Value: "",
			Usage: "expected of test case",
		},
		cli.BoolFlag{
			Name:  "seccomp",
			Usage: "whether or not enable seccomp",
		},
		cli.StringFlag{
			Name:  "cmd",
			Usage: "the command name run in container",
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
		// create container
		container := NewContainer(ctx)
		defer container.handleResult() // handle result

		if container.err != nil {
			return nil // error from NewContainer(ctx)
		}

		// input/dir/expected is required
		if container.err = checkCmdStrArgsExist(ctx, []string{"input", "dir", "expected", "cmd"}); container.err != nil {
			return nil // argument error,return nil, handle error by self
		}

		// delete cgroup
		if container.cgroup != nil {
			defer container.cgroup.Delete()
		}

		// start container
		if container.err = container.command.Start(); container.err != nil {
			return nil
		}

		// add container process to cgroup
		if container.err = container.cgroup.Add(container.command.Process.Pid); container.err != nil {
			return nil // return nil, handle error by self
		}

		// wait container exit
		container.err = container.command.Wait()
		container.done = true

		return nil // return nil, handle error by self
	},
}

type Container struct {
	id         string
	err        error
	input      string // input of test case
	baseDir    string // base directory
	memory     int64  // memory limit in kB
	timeout    int64  // time limit in ms
	scmp       bool   // enable seccomp
	expected   string // expected of test case
	cmdStr     string // command path
	command    *exec.Cmd
	stdErr     bytes.Buffer
	stdOut     bytes.Buffer
	cgroup     *cgroups.Cgroup
	waitStatus syscall.WaitStatus
	done       bool // is container done
}

func NewContainer(ctx *cli.Context) *Container {
	container := &Container{
		id:       getGlbString(ctx, "id"),
		input:    getString(ctx, "input"),
		baseDir:  getString(ctx, "dir"),
		expected: getString(ctx, "expected"),
		timeout:  ctx.Int64("timeout"),
		memory:   ctx.Int64("memory") + MemoryUsedByContainer,
		scmp:     ctx.Bool("seccomp"),
		cmdStr:   getString(ctx, "cmd"),
	}

	container.initCommand()
	if err := container.initCGroup(container.memory); err != nil {
		container.err = err // save error to container
	}

	return container
}

func (c *Container) handleResult() {
	result := model.RunResult{
		Result: model.Result{
			ID:         c.id,
			ResultType: model.RunResType,
			StdErr:     c.stdErr.String(),
		},
	}

	// get rusage and wait status info
	if c.done {
		c.waitStatus = c.command.ProcessState.Sys().(syscall.WaitStatus)
	}

	// error is not nil or exit code not 0
	if c.err != nil || c.done && c.waitStatus.ExitStatus() != 0 {
		result.Status = model.FAIL
		if c.waitStatus.ExitStatus() != 0 {
			switch c.waitStatus.Signal() {
			case syscall.SIGKILL: // oom
				result.Errno = model.OUT_OF_MEMORY
				result.Message = "out of memory"
			default: // not enough pid ,eg: fork bomb
				result.Errno = model.NO_ENOUGH_PID
				result.Message = "no enough pid"
			}
		} else {
			result.Errno = model.RUNNER_ERR
			result.Message = c.err.Error()
		}
	} else {
		// result come from container
		resStr := c.stdOut.String()
		_ = json.Unmarshal([]byte(resStr), &result) // todo There might be a bug here
	}

	res, _ := json.Marshal(result)
	fmt.Print(string(res))

	// log result
	log.WithFields(log.Fields{
		"id":       c.id,
		"input":    c.input,
		"baseDir":  c.baseDir,
		"memory":   c.memory,
		"timeout":  c.timeout,
		"scmp":     c.scmp,
		"cmdStr":   c.cmdStr,
		"expected": c.expected,
		"result":   result,
	}).Info("run result")
}

func (c *Container) initCommand() {
	args := getInitArgs()
	container := exec.Command("/proc/self/exe", args...)
	container.Args[0] = os.Args[0]
	container.Stderr = &c.stdErr
	container.Stdout = &c.stdOut
	container.Stdin = os.Stdin
	container.SysProcAttr = &syscall.SysProcAttr{
		Cloneflags: syscall.CLONE_NEWNS |
			syscall.CLONE_NEWUTS |
			syscall.CLONE_NEWIPC |
			syscall.CLONE_NEWPID |
			syscall.CLONE_NEWUSER,
		UidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getuid(),
				Size:        1,
			},
		},
		GidMappings: []syscall.SysProcIDMap{
			{
				ContainerID: 0,
				HostID:      os.Getgid(),
				Size:        1,
			},
		},
	}

	c.command = container
}

// memsw_limit: (memory + swap) limit in bytes
func (c *Container) initCGroup(memswLimit int64) error {
	var path string

	// random cgroup path
	if uid, err := uuid.NewV4(); err != nil {
		return err
	} else {
		path = "/" + uid.String()
	}

	// new cgroup
	var cpuQuota int64 = 10         // 10ms
	var kernelMem int64 = 64 * 1024 // 64m
	var disableOOMKiller = false    // kill process when oom

	// 7 pid at most
	// limit number of process to avoid fork bomb
	var pidsLimit int64 = 10
	var swappiness int64 = 0
	cgroup, err := cgroups.New(path, cgroups.Resource{
		CPU: &cgroups.CPU{
			Quota: &cpuQuota,
		},
		Memory: &cgroups.Memory{
			Limit:            &memswLimit,
			Swap:             &memswLimit,
			DisableOOMKiller: &disableOOMKiller,
			Kernel:           &kernelMem,
			Swappiness:       &swappiness,
		},
		Pids: &cgroups.Pids{
			Limit: &pidsLimit,
		},
	})
	c.cgroup = cgroup

	return err
}

func getInitArgs() []string {
	args := []string{}
	for i, arg := range os.Args {
		// append --id
		if arg == "--id" {
			args = append(args, os.Args[i], os.Args[i+1])
		}

		// replace run as init
		// append params all left
		if arg == "run" {
			params := os.Args[i:]
			params[0] = "init"
			args = append(args, params...)
			break
		}
	}

	return args
}
