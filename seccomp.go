package main

/*
#include <seccomp.h>
#include <stdint.h>

uint64_t datum(char *str) {
	return (scmp_datum_t)str;
}
*/
import "C"

import (
	"fmt"

	libseccomp "github.com/seccomp/libseccomp-golang"
)

// Syscall is a rule to match a syscall in Seccomp
type Syscall struct {
	Name       string                     `json:"name"`
	Action     libseccomp.ScmpAction      `json:"action"`
	Conditions []libseccomp.ScmpCondition `json:"conditions"`
}

// Seccomp represents syscall restrictions
// By default, only the native architecture of the kernel is allowed to be used
// for syscalls. Additional architectures can be added by specifying them in
// Architectures.
type Seccomp struct {
	DefaultAction libseccomp.ScmpAction `json:"default_action"`
	Architectures []libseccomp.ScmpArch `json:"architectures"`
	Syscalls      []*Syscall            `json:"syscalls"`
}

var seccompMap = map[string]*Seccomp{
	"blacklist-default": {
		DefaultAction: libseccomp.ActAllow,
		Architectures: []libseccomp.ScmpArch{libseccomp.ArchX86, libseccomp.ArchAMD64},
		Syscalls: []*Syscall{
			{Name: "fork", Action: libseccomp.ActKill},
			{Name: "vfork", Action: libseccomp.ActKill},
			{Name: "kill", Action: libseccomp.ActKill},
			{Name: "execveat", Action: libseccomp.ActKill},
		},
	},
}

// just allow execve when first argument is cmdStr
func notAllowExecve(config, cmdStr string) {
	fmt.Printf("address of cmdStr %p\n", &cmdStr)
	syscall := &Syscall{
		Name:   "execve",
		Action: libseccomp.ActKill,
		Conditions: []libseccomp.ScmpCondition{
			{
				Argument: 0,
				Op:       libseccomp.CompareNotEqual,
				Operand1: 78,
			},
		},
	}

	seccompMap[config].Syscalls = append(seccompMap[config].Syscalls, syscall)
}

func enableSeccomp(config, cmdStr string) (*libseccomp.ScmpFilter, error) {
	fmt.Printf("address of cmdStr %p\n", &cmdStr)
	if _, ok := seccompMap[config]; !ok {
		return nil, fmt.Errorf("libseccomp:config %s is not exist", config)
	}
	notAllowExecve(config, cmdStr)
	seccomp, _ := seccompMap[config]

	// new filter
	filter, err := libseccomp.NewFilter(seccomp.DefaultAction)
	if err != nil {
		return nil, fmt.Errorf("libseccomp:%s", err.Error())
	}

	// add arch
	for _, arch := range seccomp.Architectures {
		if err := filter.AddArch(arch); err != nil {
			return nil, fmt.Errorf("libseccomp:%s", err.Error())
		}
	}

	// set No New Privileges bit
	if err := filter.SetNoNewPrivsBit(true); err != nil {
		return nil, fmt.Errorf("libseccomp:%s", err.Error())
	}

	// add whitelist rule
	for _, syscall := range seccomp.Syscalls {
		syscallID, err := libseccomp.GetSyscallFromName(syscall.Name)
		if err != nil {
			return nil, fmt.Errorf("libseccomp:%s", err)
		}
		if err := filter.AddRuleConditional(syscallID, syscall.Action, syscall.Conditions); err != nil {
			return nil, fmt.Errorf("libseccomp:%s", err)
		}
	}

	// load
	if err := filter.Load(); err != nil {
		return nil, fmt.Errorf("libseccomp:%s", err.Error())
	}

	return filter, nil
}
