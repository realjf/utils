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
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
		return
	}
	args := []string{"-al", path}
	err = cmd.Command("ls", args...)
	if err != nil {
		log.Println(err)
		return
	}
	pid := cmd.GetPid()
	log.Printf("pid is: %d", pid)
	out, err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("%s", out)
}
