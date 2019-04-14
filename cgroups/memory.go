package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const MemoryRoot  = "/sys/fs/cgroup/memory/"

type Memory struct {
	// Memory limit (in KB).
	Limit *int64 `json:"limit,omitempty"`
	// Total memory limit (memory + swap).(in KB)
	Swap *int64 `json:"swap,omitempty"`
	// Kernel memory limit (in KB).
	Kernel *int64 `json:"kernel,omitempty"`
	// DisableOOMKiller disables the OOM killer for out of memory conditions
	DisableOOMKiller *bool `json:"disableOOMKiller,omitempty"`
}

func (m Memory)create(path string) error  {
	cgMemoryPath := filepath.Join(MemoryRoot, path)
	// create dir
	if err := os.MkdirAll(cgMemoryPath,os.ModePerm); err != nil {
		return fmt.Errorf("cgroups:%s",err.Error())
	}
	cfg := m.getConfig()

	return writeCgroupFiles(cgMemoryPath,cfg)
}

func (m Memory)delete(path string) error  {
	cgMemoryPath := filepath.Join(MemoryRoot, path)
	return deleteDir(cgMemoryPath)
}

func (m Memory)add(path string,pid int) error  {
	cgMemoryPath := filepath.Join(MemoryRoot, path)
	return addPid(cgMemoryPath,pid)
}

func (m Memory) getConfig() []cgroupFile {
	cfg := make([]cgroupFile,0)

	if m.Limit != nil {
		file := cgroupFile{
			name:"memory.limit_in_bytes" ,
			content: strconv.FormatInt(*m.Limit,10) + "K",
		}
		cfg = append(cfg,file)
	}

	if m.Swap != nil {
		file := cgroupFile{
			name:"memory.memsw.limit_in_bytes" ,
			content: strconv.FormatInt(*m.Swap,10) + "K",
		}
		cfg = append(cfg,file)
	}

	if m.Kernel != nil {
		file := cgroupFile{
			name:"memory.kmem.limit_in_bytes" ,
			content: strconv.FormatInt(*m.Kernel,10) + "K",
		}
		cfg = append(cfg,file)
	}

	if m.DisableOOMKiller != nil {
		value := 0
		if *m.DisableOOMKiller {
			value = 1
		}

		file := cgroupFile{
			name:"memory.oom_control" ,
			content: strconv.Itoa(value),
		}
		cfg = append(cfg,file)
	}

	return cfg
}
