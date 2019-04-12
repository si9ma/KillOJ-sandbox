package cgroups

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

func writeMap(dir string,fileMap map[string]string) error {
	for key, value := range fileMap {
		path := filepath.Join(dir, key)
		if err := ioutil.WriteFile(path, []byte(value), defaultPerm); err != nil {
			return fmt.Errorf("cgroups:%s",err.Error())
		}
	}

	return nil
}

func deleteDir(dir string) error {
	// return when directory not exist
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}

	if err := syscall.Rmdir(dir);err != nil {
		return fmt.Errorf("cgroups:%s",err.Error())
	}

	return nil
}

func addPid(dir string,pid int) error {
	fileName := filepath.Join(dir,"cgroup.procs")
	pidStr := strconv.Itoa(pid)
	if err := appendFile(fileName,[]byte(pidStr),defaultPerm);err != nil {
		return fmt.Errorf("cgroups:%s",err.Error())
	}

	return nil
}

func appendFile(filename string,data []byte,perm os.FileMode) error {
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	n, err := f.Write(data)
	if err == nil && n < len(data) {
		err = io.ErrShortWrite
	}
	if err1 := f.Close(); err == nil {
		err = err1
	}
	return err
}
