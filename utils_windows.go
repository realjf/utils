package utils

import (
	"strconv"

	log "github.com/sirupsen/logrus"
)

func CheckProcStateIsStopped(pid int) (bool, error) {
	args := []string{"/c", "ps", "-aeo", "ppid,pid,user,stat,pcpu,comm,wchan:32", "|", strconv.Itoa(pid), "|", "T"}
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
