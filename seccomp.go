package main

import (
	"fmt"
	"os"

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
			{Name: "execveat", Action: libseccomp.ActKill},
			{
				Name:   "open",
				Action: libseccomp.ActKill,
				Conditions: []libseccomp.ScmpCondition{
					{
						Argument: 1,
						Op:       libseccomp.CompareMaskedEqual,
						Operand1: uint64(os.O_RDWR),
						Operand2: uint64(os.O_RDWR),
					},
				},
			},
			{
				Name:   "openat",
				Action: libseccomp.ActKill,
				Conditions: []libseccomp.ScmpCondition{
					{
						Argument: 2,
						Op:       libseccomp.CompareMaskedEqual,
						Operand1: uint64(os.O_RDWR),
						Operand2: uint64(os.O_RDWR),
					},
				},
			},
			{
				Name:   "open",
				Action: libseccomp.ActKill,
				Conditions: []libseccomp.ScmpCondition{
					{
						Argument: 1,
						Op:       libseccomp.CompareMaskedEqual,
						Operand1: uint64(os.O_WRONLY),
						Operand2: uint64(os.O_WRONLY),
					},
				},
			},
			{
				Name:   "openat",
				Action: libseccomp.ActKill,
				Conditions: []libseccomp.ScmpCondition{
					{
						Argument: 2,
						Op:       libseccomp.CompareMaskedEqual,
						Operand1: uint64(os.O_WRONLY),
						Operand2: uint64(os.O_WRONLY),
					},
				},
			},
		},
	},
}

func enableSeccomp(config, cmdStr string) (*libseccomp.ScmpFilter, error) {
	if _, ok := seccompMap[config]; !ok {
		return nil, fmt.Errorf("libseccomp:config %s is not exist", config)
	}
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
