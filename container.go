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
	"time"
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
		var err error
		var result *model.RunResult
		id := ctx.GlobalString("id")

		// handle result and error by self
		defer func() {
			if err != nil {
				result = &model.RunResult{
					Result:model.Result{
						ID:id,
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
				"result": result,
			},).Info("run result")

		}()

		// input/dir/expected is required
		if err = checkCmdStrArgsExist(ctx,[]string{"input","dir","expected"});err != nil {
			return nil // return nil, handle error by self
		}

		// use cgroup to limit resource
		memory := ctx.Int64("memory")
		var cgroup *cgroups.Cgroup
		if cgroup,err = getCgroup(memory); err != nil {
			return nil // return nil, handle error by self
		}
		defer func() {
			err = cgroup.Delete()
		}()

		// start container
		container := getContainer(ctx)
		if err = container.Start(); err != nil {
			return nil
		}

		// add container process to cgroup
		if err = cgroup.Add(container.Process.Pid);err != nil {
			return nil // return nil, handle error by self
		}
		fmt.Println("container",time.Now().Nanosecond())

		// wait container exit
		err = container.Wait()

		return nil // return nil, handle error by self
	},
}

// memsw_limit: (memory + swap) limit in bytes
func getCgroup(memswLimit int64) (*cgroups.Cgroup,error) {
	var path string

	// random cgroup path
	if uid,err := uuid.NewV4();err != nil {
		return nil,err
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

	return cgroup,err
}

func getContainer(ctx *cli.Context) *exec.Cmd {
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

	return container
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
