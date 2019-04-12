package main

import (
	"fmt"
	"github.com/si9ma/KillOJ-sandbox/model"
	"github.com/urfave/cli"
	"os"
	"path/filepath"
	"syscall"
)

const hostName4Container  = "kbox"

var initCmd = cli.Command{
	Name:  "init",
	Usage: "init and run container",
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
		input := ctx.String("input")
		dir := ctx.String("dir")
		expected := ctx.String("expected")
		timeout := ctx.String("timeout")
		memory := ctx.String("memory")
		scmp := ctx.Bool("seccomp")

		// handle result and error
		defer func() {

		}()

		// init namespace
		if err = initNamespace(dir);err != nil {
			return nil // return nil, handle error by self
		}

		// when seccomp is enabled
		if scmp {
		}

		if err := syscall.Setrlimit(syscall.RLIMIT_AS,&syscall.Rlimit{Max:1,Cur:1});err != nil {
			return err
		}
		return nil
	},
}

func initNamespace(newRoot string) error {
	if err := pivotRoot(newRoot); err != nil {
		return fmt.Errorf("init namespace:%s",err.Error())
	}

	if err := syscall.Sethostname([]byte(hostName4Container)); err != nil {
		return fmt.Errorf("init namespace:%s",err.Error())
	}

	return nil
}

func pivotRoot(newRoot string) error {
	putOld := filepath.Join(newRoot, "/.pivot_root")

	// bind mount new_root to itself - this is a slight hack needed to satisfy requirement (2)
	if err := syscall.Mount(newRoot, newRoot, "", syscall.MS_BIND|syscall.MS_REC, ""); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	// create put_old directory
	if err := os.MkdirAll(putOld, 0700); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	// call pivotRoot
	if err := syscall.PivotRoot(newRoot, putOld); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	// Note that this also applies to the calling process: pivotRoot() may
	// or may not affect its current working directory.  It is therefore
	// recommended to call chdir("/") immediately after pivotRoot().
	if err := os.Chdir("/"); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	// umount put_old, which now lives at /.pivot_root
	putOld = "/.pivot_root"
	if err := syscall.Unmount(putOld, syscall.MNT_DETACH); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	// remove put_old
	if err := os.RemoveAll(putOld); err != nil {
		return fmt.Errorf("pivot_root:%s",err.Error())
	}

	return nil
}
