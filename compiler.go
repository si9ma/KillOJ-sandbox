package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/si9ma/KillOJ-sandbox/lang"
	"github.com/si9ma/KillOJ-sandbox/model"
	"github.com/urfave/cli"
	"os/exec"
	"syscall"
	"time"
)

var compileCmd = cli.Command{
	Name:  "compile",
	Usage: "compile source code",
	Description: `The compile command use to compile source code`,

	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "lang",
			Value: "",
			Usage: `language of source code`,
		},
		cli.StringFlag{
			Name:  "base-dir,dir",
			Value: "",
			Usage: "base directory of source code",
		},
		cli.StringFlag{
			Name:  "src,s",
			Value: "",
			Usage: "source code file",
		},
		cli.IntFlag{
			Name:  "timeout",
			Value: 100000,
			Usage: "compile timeout",
		},
	},
	Action: func(ctx *cli.Context) error {
		var err error
		var result *model.CompileResult

		// handle error and result
		defer func() {
			// not time limit error
			// not compile error
			if err != nil && result == nil {
				result = &model.CompileResult{
					Result: model.Result{
						ResultType: model.CompileResType,
						Status: model.FAIL,
						Errno:  model.PARMAMS_ERR,
					},
					Message: err.Error(),
				}
			}

			res,_ := json.Marshal(result)
			fmt.Println(string(res))
		}()

		// lang/dir/src is required
		if err = checkCmdStrArgsExist(ctx,[]string{"lang","dir","src"});err != nil {
			return nil // return nil, handle error by self
		}

		var compiler *exec.Cmd
		if compiler,err = getCompiler(ctx);err != nil {
			return nil // return nil, handle error by self
		}
		var stdout, stderr bytes.Buffer
		compiler.Stdout = &stdout
		compiler.Stderr = &stderr

		// compile time limit
		timeout := ctx.Int("timeout")
		time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
			// set time limit error result
			result = &model.CompileResult{
				Result: model.Result{
					ResultType: model.CompileResType,
					Status: model.FAIL,
					Errno:  model.COMPILE_TIME_LIMIT_ERR,
				},
				Message: fmt.Sprintf("compile too long(limit %dms)",timeout),
			}

			_ = syscall.Kill(-compiler.Process.Pid, syscall.SIGKILL)
		})

		if err := compiler.Run(); err != nil {
			// if result not nil, it's time limit error
			if result == nil {
				result = &model.CompileResult{
					Result: model.Result{
						ResultType: model.CompileResType,
						Status: model.FAIL,
						Errno:  model.COMPILE_ERR,
					},
					Message: stderr.String(),
				}
			}
		}

		return nil // return nil, handle error by self
	},
}

func getCompiler(ctx *cli.Context) (cmd *exec.Cmd,err error) {
	langStr := ctx.String("lang")
	baseDir := ctx.String("dir")
	src := ctx.String("src")

	if cmd,err = lang.GetCommand(langStr,src);err != nil {
		return
	}

	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	cmd.Dir = baseDir

	return
}
