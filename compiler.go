package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"syscall"
	"time"

	"github.com/si9ma/KillOJ-sandbox/model"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var COMPILER_TIMEOUT_ERR error
var COMPILER_RUN_ERR error

var compileCmd = cli.Command{
	Name:        "compile",
	Usage:       "compile source code",
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
			Value: 60000,
			Usage: "compile timeout in milliseconds",
		},
	},
	Action: func(ctx *cli.Context) error {
		compiler := NewCompiler(ctx)
		defer compiler.handleResult()

		if compiler.err != nil {
			return nil // error from NewCompiler(ctx)
		}

		// lang/dir/src is required
		if compiler.err = checkCmdStrArgsExist(ctx, []string{"lang", "dir", "src"}); compiler.err != nil {
			return nil // return nil, handle error by self
		}

		// compile time limit
		time.AfterFunc(time.Duration(compiler.timeout)*time.Millisecond, func() {
			if compiler.cmd.Process != nil {
				COMPILER_TIMEOUT_ERR = fmt.Errorf("compile too long(limit %dms)", compiler.timeout)
				compiler.err = COMPILER_TIMEOUT_ERR
				_ = syscall.Kill(-compiler.cmd.Process.Pid, syscall.SIGKILL)
			}
		})

		if err := compiler.cmd.Run(); err != nil {
			COMPILER_RUN_ERR = err
			compiler.err = COMPILER_RUN_ERR
		}

		return nil // return nil, handle error by self
	},
}

type Lang struct {
	Command    string
	SubCommand string
	Args       []string
}

var langCompilerMap = map[string]Lang{
	"c": {
		Command: "/usr/bin/gcc",
		Args:    []string{"-save-temps", "-std=c11", "-fmax-errors=10", "-static", "-o", "Main"},
	},
	"cpp": {
		Command: "/usr/bin/g++",
		Args:    []string{"-save-temps", "-std=c++11", "-fmax-errors=10", "-static", "-o", "Main"},
	},
	"go": {
		Command:    "/usr/bin/go",
		SubCommand: "build",
		Args:       []string{"-o", "Main"},
	},
	"java": {
		Command: "/usr/bin/javac",
		Args:    []string{"-J-Xmx500m", "-Xmaxwarns", "10", "-Xmaxerrs", "10"}, // heap memory limit to 500m, max error and warn limit to 10
	},
}

type Compiler struct {
	id      string
	lang    string
	err     error
	cmd     *exec.Cmd // compiler cmd
	baseDir string
	src     string // source code file name
	stdOut  bytes.Buffer
	stdErr  bytes.Buffer
	timeout int64 // timeout in ms
}

func NewCompiler(ctx *cli.Context) *Compiler {
	compiler := &Compiler{
		id:      ctx.String("id"),
		lang:    ctx.String("lang"),
		baseDir: ctx.String("dir"),
		src:     ctx.String("src"),
		timeout: ctx.Int64("timeout"),
	}

	if lang, ok := langCompilerMap[compiler.lang]; ok {
		params := append(lang.Args, compiler.src)
		// check if need sub cmd(eg: go build)
		if lang.SubCommand != "" {
			params = append([]string{lang.SubCommand}, params...)
		}
		compiler.cmd = exec.Command(lang.Command, params...)
	} else {
		compiler.err = fmt.Errorf("language %s is not supported", compiler.lang)
		return compiler
	}

	compiler.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	compiler.cmd.Dir = compiler.baseDir
	compiler.cmd.Stdout = &compiler.stdOut
	compiler.cmd.Stderr = &compiler.stdErr

	return compiler
}

func (c *Compiler) handleResult() {
	result := &model.CompileResult{
		Result: model.Result{
			ID:         c.id,
			ResultType: model.CompileResType,
			StdErr:     c.stdErr.String(),
		},
	}

	// handle error
	if c.err != nil {
		result.Message = c.err.Error()
		result.Status = model.FAIL

		switch c.err {
		case COMPILER_TIMEOUT_ERR:
			result.Errno = model.COMPILE_TIMEOUT
		case COMPILER_RUN_ERR:
			result.Errno = model.INNER_COMPILER_ERR
		default:
			result.Errno = model.OUTER_COMPILER_ERR
		}
	} else {
		result.Status = model.SUCCESS
		result.Message = "compile success"
	}

	res, _ := json.Marshal(result)
	fmt.Println(string(res))

	// log result
	log.WithFields(log.Fields{
		"id":      c.id,
		"lang":    c.lang,
		"src":     c.src,
		"baseDir": c.baseDir,
		"timeout": c.timeout,
		"result":  result,
	}).Info("compile result")
}
