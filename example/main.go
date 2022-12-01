package main

import (
	"log"
	"os"
	"os/user"

	"github.com/realjf/utils"
)

func main() {
	cmd := utils.NewCmd()
	user, err := user.Current()
	if err != nil {
		log.Println(err)
		return
	}
	cmd.SetUser(user)
	defer cmd.Close()
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	// cmd.SetTimeout(1 * time.Second)
	args := []string{"-al", path}
	pid, err := cmd.Command("ls", args...)
	if err != nil {
		log.Println(err)
		return
	}

	cmd.NeedInput("hello:")
	pid1 := cmd.GetPid()
	log.Printf("pid is: %d, %d", pid, pid1)
	out, err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("%s", out)
}
