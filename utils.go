package utils

import (
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

func IsExist(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	} else {
		return true
	}
}

func CheckProcStateIsStopped(pid int) (bool, error) {
	args := []string{"-aeo", "ppid,pid,user,stat,pcpu,comm,wchan:32", "|", strconv.Itoa(pid), "|", "T"}
	out, err := NewCmd().RunCommand("ps", args...)
	if err != nil {
		return false, err
	}
	log.Infof("%s", out)
	if string(out) != "" {
		return true, nil
	}
	return false, nil
}
