package main

import (
	"bytes"
	"fmt"
	"github.com/si9ma/KillOJ-sandbox/lang"
	"github.com/urfave/cli"
	"os"
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

		// lang/dir/src is required
		if err := checkCmdStrArgsExist(ctx,[]string{"lang","dir","src"});err != nil {
			return err
		}

		langStr := ctx.String("lang")
		baseDir := ctx.String("dir")
		src := ctx.String("src")
		timeout := ctx.Int("timeout")
		var cmd *exec.Cmd

		if cmd,err = lang.GetCommand(langStr,src);err != nil {
			return err
		}

		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		cmd.Dir = baseDir

		time.AfterFunc(time.Duration(timeout)*time.Millisecond, func() {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		})

		if err := cmd.Run(); err != nil {
			_, _ = os.Stderr.WriteString(fmt.Sprintf("stderr: %s, err: %s\n", stderr.String(), err.Error()))
			return err
		}

		_, _ = os.Stdout.WriteString("Compile OK\n")

		return err
	},
}

