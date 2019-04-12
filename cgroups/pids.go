package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const PidsRoot = "/sys/fs/cgroup/pids/"

type Pids struct {
	// Maximum number of PIDs. Default is "no limit".
	Limit *int64 `json:"limit"`
}

func (p Pids)create(path string) error  {
	cgPidsPath := filepath.Join(PidsRoot, path)
	// create dir
	if err := os.MkdirAll(cgPidsPath,os.ModePerm); err != nil {
		return fmt.Errorf("cgroups:%s",err.Error())
	}
	mapping := p.getMapping()

	return writeMap(cgPidsPath,mapping)
}

func (p Pids)delete(path string) error  {
	cgPidsPath := filepath.Join(PidsRoot, path)
	return deleteDir(cgPidsPath)
}

func (p Pids)add(path string,pid int) error  {
	cgPidsPath := filepath.Join(PidsRoot, path)
	return addPid(cgPidsPath,pid)
}

func (p Pids)getMapping() map[string]string {
	mapping := make(map[string]string)

	if p.Limit != nil {
		mapping["pids.max"] = strconv.FormatInt(*p.Limit,10)
	}

	return mapping
}

