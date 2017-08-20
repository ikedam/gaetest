package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path"
	"strings"
	"syscall"
	"time"
)

func runCommand(name string, args... string) int {
	fmt.Fprintf(os.Stderr, "# %v %v\n", name, strings.Join(args, " "))
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

func shutdownServer(cmd *exec.Cmd) bool {
	if resp, err := http.Get("http://localhost:8000/quit"); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop server: %v", err)
		return false
	} else if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Failed to stop server: %v", resp)
		return false
	}
	return true
}

func pushd(dir string) func() {
	orig, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	if err := os.Chdir(dir); err != nil {
		panic(err)
	}
	return func() {
		if err := os.Chdir(orig); err != nil {
			panic(err)
		}
	}
}

func runYarn() int {
	popd := pushd("client")
	defer popd()
	return runCommand("yarn")
}

func runAngularUnitTest(isDocker bool) int {
	popd := pushd("client")
	defer popd()

	args := []string {
		"run",
		"test",
		"--",
		"--single-run",
		"true",
	}
	if isDocker {
		args = append(args, "--config", "karma-docker.conf.js")
	}

	return runCommand("yarn", args...)
}

func runE2ETest(isDocker bool) int {
	serverStdout, err := os.OpenFile(
		"server.stdout.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer serverStdout.Close()
	serverStderr, err := os.OpenFile(
		"server.stderr.log",
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer serverStderr.Close()

	python, err := exec.LookPath("python")
	if err != nil {
		panic(err)
	}
	devAppServer, err := exec.LookPath("dev_appserver.py")
	if err != nil {
		panic(err)
	}
	args := []string {
		devAppServer,
		"--enable_watching_go_path=false",
		"--clear_datastore=true",
		"--datastore_consistency_policy=consistent",
		"--automatic_restart=false",
		"--watcher_ignore_re=.*",
		"--skip_sdk_update_check=true",
		"server/app.yaml",
	}

	fmt.Fprintf(os.Stderr, "# %v %v\n", python, strings.Join(args, " "))
	cmd := exec.Command(python, args...)
	cmd.Stdout = serverStdout
	cmd.Stderr = serverStderr
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	defer shutdownServer(cmd)

	url := "http://localhost:8080/"
	seconds := 30
	fmt.Fprintf(os.Stderr, "Wait for server gets ready for %v seconds: %v\n", seconds, url)
	limit := time.Now().Add(time.Duration(seconds) * time.Second)
	client := &http.Client {
		Timeout: time.Duration(seconds) * time.Second,
	}
	for true {
		if time.Since(limit) > 0 {
			panic(fmt.Sprintf("Server didnot get ready after %v seconds", seconds))
		}
		if _, err := client.Get(url); err == nil {
			break
		}
		time.Sleep(time.Duration(100) * time.Millisecond)
	}

	popd := pushd("client")
	defer popd()

	e2eArgs := []string {
		"run",
		"e2e",
		"--", 
		"--serve",
		"true",
		"--port",
		"4200",
	}
	if isDocker {
		e2eArgs = append(e2eArgs, "--config", "protractor-docker.conf.js")
	}

	return runCommand("yarn", e2eArgs...)
}

func main() {
	isDocker := flag.Bool("docker", false, "Specify when running in docker")
	flag.Parse()

	gopath := os.Getenv("GOPATH")
	projectDir := path.Join(gopath, "src", "github.com", "ikedam", "gaetest")
	if err := os.Chdir(projectDir); err != nil {
		panic(err)
	}
	if ret := runCommand("glide", "install"); ret != 0 {
		os.Exit(ret)
	}
	if ret := runCommand("goapp", "test", "./server", "./testutil"); ret != 0 {
		os.Exit(ret)
	}
	if ret := runYarn(); ret != 0 {
		os.Exit(ret)
	}
	if ret := runAngularUnitTest(*isDocker); ret != 0 {
		os.Exit(ret)
	}
	if ret := runE2ETest(*isDocker); ret != 0 {
		os.Exit(ret)
	}
	os.Exit(0)
}
