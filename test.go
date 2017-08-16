package main

import (
	"os"
	"os/exec"
	"path"
	"syscall"
)

func runCommand(name string, args... string) int {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			s := exitErr.Sys()
			if waitStatus, ok := s.(*syscall.WaitStatus); ok {
				return waitStatus.ExitStatus()
			}
		}
		panic(err)
	}
	return 0
}

func main() {
	gopath := os.Getenv("GOPATH")
	projectDir := path.Join(gopath, "src", "github.com", "ikedam", "gaetest")
	if err := os.Chdir(projectDir); err != nil {
		panic(err)
	}
	if ret := runCommand("glide", "install"); ret != 0 {
		os.Exit(ret)
	}
	os.Exit(runCommand("goapp", "test", "./server", "./testutil"))
}
