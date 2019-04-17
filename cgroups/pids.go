package cgroups

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

const PidsRoot = "/sys/fs/cgroup/pids/"

type Pids struct {
	Limit *int64 `json:"limit"` // Maximum number of PIDs. Default is "no limit".
}

func (p Pids) create(path string) error {
	cgPidsPath := filepath.Join(PidsRoot, path)
	// create dir
	if err := os.MkdirAll(cgPidsPath, os.ModePerm); err != nil {
		return fmt.Errorf("cgroups:%s", err.Error())
	}
	cfg := p.getConfig()

	return writeCgroupFiles(cgPidsPath, cfg)
}

func (p Pids) delete(path string) error {
	cgPidsPath := filepath.Join(PidsRoot, path)
	return deleteDir(cgPidsPath)
}

func (p Pids) add(path string, pid int) error {
	cgPidsPath := filepath.Join(PidsRoot, path)
	return addPid(cgPidsPath, pid)
}

func (p Pids) getConfig() []cgroupFile {
	cfg := make([]cgroupFile, 0)

	if p.Limit != nil {
		file := cgroupFile{
			name:    "pids.max",
			content: strconv.FormatInt(*p.Limit, 10),
		}
		cfg = append(cfg, file)
	}

	return cfg
}
