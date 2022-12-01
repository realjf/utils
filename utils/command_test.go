package utils

import (
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
	args := []string{"-al", "$(pwd)"}
	err = cmd.Command("ls", args...)
	if err != nil {
		t.Fatalf(err.Error())
	}
	pid := cmd.GetPid()
	t.Logf("%d", pid)
	out, err := cmd.Run()
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("%s", out)
}
