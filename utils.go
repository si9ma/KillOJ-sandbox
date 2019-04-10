package main

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
)

// fatal prints the error's details
// then exits the program with an exit status of 1.
func fatal(err error) {
	// make sure the error is written to the logger
	logrus.Error(err)
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

// check if arguments exist for command
func checkCmdStrArgsExist(ctx *cli.Context,args []string) (err error) {
	cmdName := ctx.Command.Name

	for _,arg := range args {
		if ctx.String(arg) == "" {
			err = fmt.Errorf("%s: %q require parameter:--%s", os.Args[0], cmdName,arg)
		}
	}

	return err
}