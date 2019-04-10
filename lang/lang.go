package lang

import (
	"fmt"
	"os/exec"
)

type Lang struct {
	Command string
	SubCommand string
	Args []string
}

var langMap = map[string]Lang {
	"c": {
		Command: "/usr/bin/gcc",
		Args:[]string{"-save-temps", "-std=c11", "-fmax-errors=10", "-static", "-o", "Main"},
	},
	"cpp": {
		Command: "/usr/bin/g++",
		Args:[]string{"-save-temps", "-std=c++11", "-fmax-errors=10", "-static", "-o", "Main"},
	},
	"go": {
		Command: "/usr/bin/go",
		SubCommand: "build",
		Args: []string{"-o","Main"},
	},
}

// get command by language
func GetCommand(langStr,src string) (cmd *exec.Cmd,err error)  {
	if lang,ok := langMap[langStr];ok {
		params := append(lang.Args,src)
		// check if need sub command(eg: go build)
		if lang.SubCommand != "" {
			params = append([]string{lang.SubCommand},params...)
		}
		cmd = exec.Command(lang.Command,params...)
	}else {
		err = fmt.Errorf("language %s is not supported",langStr)
	}

	return
}

