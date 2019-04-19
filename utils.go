package main

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
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
func checkCmdStrArgsExist(ctx *cli.Context, args []string) (err error) {
	cmdName := ctx.Command.Name

	for _, arg := range args {
		if ctx.String(arg) == "" {
			err = fmt.Errorf("%s: %q require parameter:--%s", os.Args[0], cmdName, arg)
		}
	}

	return err
}

// command line flag treat \n as two character(\ and n)
// escape string by self
// just escape \" \' \v \r \f \n \t \\ \a \b
func escapeString(src string) string {
	var escapeMap = map[byte]byte{
		'"':  '"',
		'\'': '\'',
		'a':  '\a',
		'b':  '\b',
		'f':  '\f',
		'n':  '\n',
		'v':  '\v',
		't':  '\t',
		'r':  '\r',
		'\\': '\\',
	}

	dst := []byte(src)
	d := 0
	for s := 0; s < len(src); s++ {
		if src[s] == '\\' && s < len(src) {
			if rel, ok := escapeMap[src[s+1]]; ok {
				dst[d] = rel
				s++ // skip next character
			}
		}
		d++
	}

	return string(dst[:d])
}

// command line flag treat \n as two character(\ and n)
// escape string by self
func getString(ctx *cli.Context, name string) string {
	return escapeString(ctx.String(name))
}

// command line flag treat \n as two character(\ and n)
// escape string by self
func getGlbString(ctx *cli.Context, name string) string {
	return escapeString(ctx.GlobalString(name))
}
