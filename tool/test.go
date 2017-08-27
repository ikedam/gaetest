package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"
)

type testOption struct{
	IsVerbose bool
	IsDocker bool
}

type testCommandInfo struct{
	Command string
	Func func(testOption) int
	NeedGlide bool
	NeedYarn bool
	Description string
}

var testCommands = []*testCommandInfo{
	&testCommandInfo{
		Command: "sunit",
		Func: runGoUnitTest,
		NeedGlide: true,
		NeedYarn: false,
		Description: "Run goapp test",
	},
	&testCommandInfo{
		Command: "cunit",
		Func: runAngularUnitTest,
		NeedGlide: false,
		NeedYarn: true,
		Description: "Run ng test",
	},
	&testCommandInfo{
		Command: "e2e",
		Func: runE2ETest,
		NeedGlide: true,
		NeedYarn: true,
		Description: "Run ng e2e",
	},
}

func main() {
	// os.Args[0] points to the compiled binary
	_, arg0, _, _ := runtime.Caller(0)
	parser := flag.NewFlagSet(arg0, flag.ExitOnError)
	parser.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [Options] [Commands]:\n", arg0)
		fmt.Fprintf(os.Stderr, "Options:\n")
		parser.PrintDefaults()
		fmt.Fprintf(os.Stderr, "Commands:\n")
		for _, c := range(testCommands){
			fmt.Fprintf(os.Stderr, "  %s: %s\n", c.Command, c.Description)
		}
	}
	option := testOption{}
	parser.BoolVar(&option.IsVerbose, "v", false, "Verbose output")
	parser.BoolVar(&option.IsVerbose, "verbose", false, "Verbose output")
	parser.BoolVar(&option.IsDocker, "docker", false, "Specify when running in docker")
	noSetup := parser.Bool("nosetup", false, "Skip calling glide install or yarn")
	parser.Parse(os.Args[1:])
	commands := parser.Args()
	if len(commands) <= 0 {
		for _, c := range(testCommands){
			commands = append(commands, c.Command)
		}
	}

	var needGlide bool
	var needYarn bool
	commandsToRun := []*testCommandInfo{}
	commandMap := map[string]*testCommandInfo{}
	for _, c := range(testCommands){
		commandMap[c.Command] = c
	}
	for _, command := range(commands) {
		c, ok := commandMap[command]
		if !ok {
			fmt.Fprintf(os.Stderr, "Unknown Command: %s\n", command)
			flag.Usage()
			os.Exit(2)
		}
		commandsToRun = append(commandsToRun, c)
		needGlide = needGlide || c.NeedGlide
		needYarn = needYarn || c.NeedYarn
	}

	// locate the project directory
	// this file should be located at projectdir/tool/
	projectDir := filepath.Clean(filepath.Join(filepath.Dir(arg0), ".."))
	if err := os.Chdir(projectDir); err != nil {
		panic(err)
	}

	if needGlide && !*noSetup {
		if ret := runCommand("glide", "install"); ret != 0 {
			os.Exit(ret)
		}
	}
	if needYarn && !*noSetup {
		func(){
			popd := pushd("client")
			defer popd()
			if ret := runCommand("yarn"); ret != 0 {
				os.Exit(ret)
			}
		}()
	}
	for _, c := range(commandsToRun){
		if ret := c.Func(option); ret != 0 {
			os.Exit(ret)
		}
	}
	os.Exit(0)
}

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

func runGoUnitTest(option testOption) int {
	args := []string{
		"test",
		"./server",
		"./testutil",
	}
	if option.IsVerbose {
		args = append(args, "-v")
	}
	return runCommand("goapp", args...)
}

func runAngularUnitTest(option testOption) int {
	popd := pushd("client")
	defer popd()

	args := []string {
		"run",
		"test",
		"--",
		"--single-run",
		"true",
	}
	if option.IsDocker {
		args = append(args, "--config", "karma-docker.conf.js")
	}

	return runCommand("yarn", args...)
}

func runE2ETest(option testOption) int {
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
		"--require_indexes=true",
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
	if option.IsDocker {
		e2eArgs = append(e2eArgs, "--config", "protractor-docker.conf.js")
	}

	return runCommand("yarn", e2eArgs...)
}
