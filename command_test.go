package utils

import (
	"log"
	"os"
	"os/user"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := NewCmd()
	user, err := user.Current()
	if err != nil {
		t.Fatalf(err.Error())
	}
	cmd.SetUser(user)
	defer cmd.Close()

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	args := []string{"-al", path}
	pid, err := cmd.Command("ls", args...)
	if err != nil {
		t.Fatalf(err.Error())
	}
	cmd.NeedInput("hello:")
	pid1 := cmd.GetPid()
	t.Logf("%d,%d", pid, pid1)
	out, err := cmd.Run()
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%s", out)
}
