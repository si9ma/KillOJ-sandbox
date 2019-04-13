package main

import (
	"encoding/json"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/si9ma/KillOJ-sandbox/cgroups"
	"github.com/si9ma/KillOJ-sandbox/model"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"syscall"
)

var runCmd = cli.Command{
	Name:  "run",
	Usage: "run app in container",
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
			Name: "seccomp",
			Usage: "whether or not enable seccomp",
		},
		cli.StringFlag{
			Name: "cmd",
			Usage: "the command name run in container",
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
		// create container
		container := NewContainer(ctx)
		if container.err != nil {
			return nil
		}

		// handle result
		container.handleResult()

		// delete cgroup
		defer container.cgroup.Delete()

		// input/dir/expected is required
		if container.err = checkCmdStrArgsExist(ctx,[]string{"input","dir","expected","cmd"});container.err != nil {
			return nil // return nil, handle error by self
		}

		// start container
		if container.err = container.command.Start(); container.err != nil {
			return nil
		}

		// add container process to cgroup
		if container.err = container.cgroup.Add(container.command.Process.Pid);container.err != nil {
			return nil // return nil, handle error by self
		}

		// wait container exit
		container.err = container.command.Wait()

		return nil // return nil, handle error by self
	},
}

type Container struct {
	id string
	err error
	input string // input of test case
	baseDir string // base directory
	memory int64 // memory limit in kB
	timeout int64 // time limit in ms
	scmp bool // enable seccomp
	expected string // expected of test case
	cmdStr string // command path
	command *exec.Cmd
	cgroup *cgroups.Cgroup
}

func NewContainer(ctx *cli.Context) *Container {
	container := &Container{
		id: ctx.GlobalString("id"),
		input: ctx.String("input"),
		baseDir: ctx.String("dir"),
		expected: ctx.String("expected"),
		timeout: ctx.Int64("timeout"),
		memory: ctx.Int64("memory"),
		scmp: ctx.Bool("seccomp"),
		cmdStr: ctx.String("cmd"),
	}

	container.initCommand()
	if err := container.initCGroup(container.memory); err != nil {
		container.err = err // save error to container
	}

	return container
}

func (c *Container)handleResult()  {
	if c.err != nil {
		result := &model.RunResult{
			Result:model.Result{
				ID:c.id,
				ResultType: model.RunResType,
				Status: model.FAIL,
				Errno: model.RUNNER_ERR,
				Message: err.Error(),
			},
		}
	}

	res,_ := json.Marshal(result)
	fmt.Println(string(res))

	// log result
	log.WithFields(log.Fields{
		"id": ctx.GlobalString("id"),
		"input": ctx.String("input"),
		"baseDir": ctx.String("dir"),
		"memory": ctx.String("memory"),
		"timeout": ctx.String("timeout"),
		"scmp": ctx.String("scmp"),
		"cmdStr": ctx.String("cmdStr"),
		"result": result,
	},).Info("run result")
}

func (c *Container)initCommand() {
	args := getInitArgs()
	container := exec.Command("/proc/self/exe",args...)
	container.Args[0] = os.Args[0]
	container.Stdout = os.Stdout
	container.Stderr = os.Stderr
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
func (c *Container)initCGroup(memswLimit int64) error {
	var path string

	// random cgroup path
	if uid,err := uuid.NewV4();err != nil {
		return err
	}else {
		path = "/" + uid.String()
	}

	// new cgroup
	var cpuQuota int64 = 10 // 10ms
	var kernelMem int64 = 64000 // 64m
	var disableOOMKiller bool = false // kill process when oom
	var pidsLimit int64 = 64
	cgroup,err := cgroups.New(path,cgroups.Resource{
		CPU: &cgroups.CPU{
			Quota: &cpuQuota,
		},
		Memory: &cgroups.Memory{
			Limit: &memswLimit,
			Swap: &memswLimit,
			DisableOOMKiller: &disableOOMKiller,
			Kernel: &kernelMem,
		},
		Pids: &cgroups.Pids{
			Limit: &pidsLimit,
		},
	})
	c.cgroup = cgroup

	return err
}

func getInitArgs() []string {
	args := os.Args
	for i,arg := range args {
		// replace command
		if arg == "run" {
			args[i] = "init"
		}
	}

	// remove args[0]
	return args[1:]
}
