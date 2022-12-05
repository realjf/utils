package utils

import (
	"log"
	"testing"
)

func TestCmd(t *testing.T) {
	cmd := NewCmd()
	// user, err := user.Current()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetUser(user)
	defer cmd.Close()
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetTimeout(1 * time.Second)
	args := []string{"--cpu", "1", "--vm", "1", "--vm-bytes", "220M", "--timeout", "10s", "--vm-keep"}
	pid, err := cmd.Command("stress", args...)
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

func TestCmd2(t *testing.T) {
	cmd := NewCmd().SetDebug(true)
	// user, err := user.Current()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetUser(user)
	defer cmd.Close()
	// path, err := os.Getwd()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	// cmd.SetTimeout(1 * time.Second)
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
