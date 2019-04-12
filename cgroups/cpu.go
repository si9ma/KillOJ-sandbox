package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const CPURoot  = "/sys/fs/cgroup/cpu/"

type CPU struct {
	// CPU hardcap limit (in usecs). Allowed cpu time in a given period.(ms)
	Quota *int64 `json:"quota,omitempty"`
}

func (c CPU)create(path string) error  {
	cgCPUPath := filepath.Join(CPURoot, path)
	// create dir
	if err := os.MkdirAll(cgCPUPath,os.ModePerm); err != nil {
		return fmt.Errorf("cgroups:%s",err.Error())
	}
	mapping := c.getMapping()

	return writeMap(cgCPUPath,mapping)
}

func (c CPU)delete(path string) error  {
	cgCPUPath := filepath.Join(CPURoot, path)
	return deleteDir(cgCPUPath)
}

func (c CPU)add(path string,pid int) error  {
	cgCPUPath := filepath.Join(CPURoot, path)
	return addPid(cgCPUPath,pid)
}

func (c CPU)getMapping() map[string]string {
	mapping := make(map[string]string)

	if c.Quota != nil {
		mapping["cpu.cfs_quota_us"] = strconv.FormatInt(*c.Quota * 1000,10)
	}

	return mapping
}