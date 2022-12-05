package main

import (
	"log"
	"os"

	"github.com/realjf/utils"
)

// running under sudo
func main() {
	cmd := utils.NewCmd().SetDebug(true)
	cmd.SetUsername(os.Getenv("SUDO_USER"))
	defer cmd.Close()
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetTimeout(1 * time.Second)

	// attr := syscall.SysProcAttr{
	// 	// Cloneflags:                 syscall.CLONE_NEWUTS | syscall.CLONE_NEWIPC | syscall.CLONE_NEWPID | syscall.CLONE_NEWNS | syscall.CLONE_NEWUSER | syscall.CLONE_NEWNET,
	// 	// GidMappingsEnableSetgroups: true,
	// 	Setpgid: true,
	// 	// UidMappings: []syscall.SysProcIDMap{
	// 	// 	{
	// 	// 		ContainerID: 0,
	// 	// 		HostID:      0,
	// 	// 		Size:        1,
	// 	// 	},
	// 	// },
	// 	// GidMappings: []syscall.SysProcIDMap{
	// 	// 	{
	// 	// 		ContainerID: 0,
	// 	// 		HostID:      0,
	// 	// 		Size:        1,
	// 	// 	},
	// 	// },
	// 	Pgid:       cmd.GetGid(),
	// 	Credential: &syscall.Credential{},
	// }
	// cmd.SetSysProcAttr(attr)
	// user, err := user.Current()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetUser(user)

	cmd.SetNoSetGroups(true)

	args := []string{"/home/realjf/Downloads/e780530a-5eac-4118-9aa0-cb2d2f3e7db8.epub", "/home/realjf/Downloads/e780530a-5eac-4118-9aa0-cb2d2f3e7db8.pdf"}
	pid, err := cmd.Command("ebook-convert", args...)
	if err != nil {
		log.Println(err)
		return
	}

	// cmd.NeedInput("hello:")
	pid1 := cmd.GetPid()
	log.Printf("pid is: %d, %d", pid, pid1)
	out, err := cmd.Run()
	if err != nil {
		log.Println(err.Error())
		return
	}
	log.Printf("%s", out)
}
