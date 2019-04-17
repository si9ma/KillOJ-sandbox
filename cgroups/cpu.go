package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const CPURoot = "/sys/fs/cgroup/cpu/"

type CPU struct {
	Quota *int64 `json:"quota,omitempty"` // CPU hardcap limit (in usecs). Allowed cpu time in a given period.(ms)
}

func (c CPU) create(path string) error {
	cgCPUPath := filepath.Join(CPURoot, path)
	// create dir
	if err := os.MkdirAll(cgCPUPath, os.ModePerm); err != nil {
		return fmt.Errorf("cgroups:%s", err.Error())
	}
	cfg := c.getConfig()

	return writeCgroupFiles(cgCPUPath, cfg)
}

func (c CPU) delete(path string) error {
	cgCPUPath := filepath.Join(CPURoot, path)
	return deleteDir(cgCPUPath)
}

func (c CPU) add(path string, pid int) error {
	cgCPUPath := filepath.Join(CPURoot, path)
	return addPid(cgCPUPath, pid)
}

func (c CPU) getConfig() []cgroupFile {
	cfg := make([]cgroupFile, 0)

	if c.Quota != nil {
		file := cgroupFile{
			name:    "cpu.cfs_quota_us",
			content: strconv.FormatInt(*c.Quota*1000, 10),
		}
		cfg = append(cfg, file)
	}

	return cfg
}
