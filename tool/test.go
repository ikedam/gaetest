package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"time"
)

type testOption struct {
	IsVerbose  bool
	IsDocker   bool
	IsCoverage bool
	LogDir     string
}

type testCommandInfo struct {
	Command     string
	Func        func(testOption) int
	NeedGlide   bool
	NeedYarn    bool
	Description string
}

var testCommands = []*testCommandInfo{
	&testCommandInfo{
		Command:     "sunit",
		Func:        runGoUnitTest,
		NeedGlide:   true,
		NeedYarn:    false,
		Description: "Run goapp test",
	},
	&testCommandInfo{
		Command:     "cunit",
		Func:        runAngularUnitTest,
		NeedGlide:   false,
		NeedYarn:    true,
		Description: "Run ng test",
	},
	&testCommandInfo{
		Command:     "e2e",
		Func:        runE2ETest,
		NeedGlide:   true,
		NeedYarn:    true,
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
		for _, c := range testCommands {
			fmt.Fprintf(os.Stderr, "  %s: %s\n", c.Command, c.Description)
		}
	}
	option := testOption{}
	parser.BoolVar(&option.IsVerbose, "v", false, "Verbose output")
	parser.BoolVar(&option.IsVerbose, "verbose", false, "Verbose output")
	parser.BoolVar(&option.IsDocker, "docker", false, "Specify when running in docker")
	parser.BoolVar(&option.IsCoverage, "coverage", false, "Output coverage")
	parser.StringVar(&option.LogDir, "log", ".", "Specify the directory to output logs")
	noSetup := parser.Bool("nosetup", false, "Skip calling glide install or yarn")
	parser.Parse(os.Args[1:])
	commands := parser.Args()
	if len(commands) <= 0 {
		for _, c := range testCommands {
			commands = append(commands, c.Command)
		}
	}

	var needGlide bool
	var needYarn bool
	commandsToRun := []*testCommandInfo{}
	commandMap := map[string]*testCommandInfo{}
	for _, c := range testCommands {
		commandMap[c.Command] = c
	}
	for _, command := range commands {
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
		func() {
			popd := pushd("client")
			defer popd()
			if ret := runCommand("yarn"); ret != 0 {
				os.Exit(ret)
			}
		}()
	}
	for _, c := range commandsToRun {
		if ret := c.Func(option); ret != 0 {
			os.Exit(ret)
		}
	}
	os.Exit(0)
}

func runCommand(name string, args ...string) int {
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

// カバレッジファイルの出力フォーマットは
// mode: set
// file.go:line.column,line.column statements count
// file.go:line.column,line.column statements count
// ...
// というフォーマットになっている。
// count >= 1 であれば実行されているということなので、
// * 逆順ソートする
// * 前方 2 カラムが同一であれば 2 行目移行は無視
// という処理をする。
type coverageSummarizer struct {
	coverageLogs []string
}

func (coverage *coverageSummarizer) Load(path string) error {
	var scanner *bufio.Scanner
	if temp, err := os.Open(path); err == nil {
		defer temp.Close()
		scanner = bufio.NewScanner(temp)
	} else {
		return err
	}
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "mode:") {
			continue
		}
		coverage.coverageLogs = append(coverage.coverageLogs, line)
	}
	return scanner.Err()
}

func (coverage *coverageSummarizer) Save(path string) error {
	sort.Sort(sort.Reverse(sort.StringSlice(coverage.coverageLogs)))
	var coverageFile *bufio.Writer
	if _coverageFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644); err == nil {
		coverageFile = bufio.NewWriter(_coverageFile)
		defer _coverageFile.Close()
	} else {
		return err
	}
	if _, err := coverageFile.WriteString("mode: set\n"); err != nil {
		return err
	}
	label := ""
	for _, line := range coverage.coverageLogs {
		// 範囲 行数 実行回数
		// になっている。原則、範囲 行数 はいつも同じ値になるので、
		// 実行回数が最大のものだけ （＝ソート順で最大のものだけ） 採用する。
		units := strings.Split(line, " ")
		if len(units) != 3 {
			log.Printf("Ignore unexpected line: %v", line)
			continue
		}
		currentLabel := strings.Join(units[0:2], " ")
		if label == currentLabel {
			continue
		}
		label = currentLabel
		if _, err := coverageFile.WriteString(fmt.Sprintf("%s\n", line)); err != nil {
			return err
		}
	}
	return coverageFile.Flush()
}

func runGoUnitTest(option testOption) int {
	packages := []string{
		"./server",
		"./testutil",
	}
	args := []string{
		"test",
	}
	if option.IsVerbose {
		args = append(args, "-v")
	}
	if !option.IsCoverage {
		// カバレッジを使用しない場合、複数のパッケージのテストを一度に起動して良い。
		args = append(args, packages...)
		return runCommand("goapp", args...)
	}

	// カバレッジを使用する場合、パッケージごとに実行が必要
	ret := 0
	coverage := &coverageSummarizer{}
	for idx, packageName := range packages {
		if !func() bool {
			tempCoverageFilePath := filepath.Join(
				option.LogDir,
				fmt.Sprintf("coverage_tmp_%02d.out", idx),
			)
			argsForCoverage := make([]string, len(args))
			copy(argsForCoverage, args)
			argsForCoverage = append(
				argsForCoverage,
				fmt.Sprintf("-coverprofile=%s", tempCoverageFilePath),
				fmt.Sprintf("-coverpkg=%s", strings.Join(packages, ",")),
				packageName,
			)
			ret = runCommand("goapp", argsForCoverage...)
			if ret != 0 {
				return false
			}
			defer os.Remove(tempCoverageFilePath)

			if err := coverage.Load(tempCoverageFilePath); err != nil {
				log.Printf("ERROR: Failed to read from %v: %v", tempCoverageFilePath, err)
				ret = 1
				return false
			}

			return true
		}() {
			break
		}
	}

	coverageFilePath := filepath.Join(option.LogDir, "coverage.out")
	htmlFilePath := filepath.Join(option.LogDir, "coverage.html")
	if err := coverage.Save(coverageFilePath); err != nil {
		log.Printf("ERROR: Failed to write: %v: %v", coverageFilePath, err)
		ret = 1
	} else {
		if htmlRet := runCommand(
			"goapp",
			"tool",
			"cover",
			fmt.Sprintf("-html=%s", coverageFilePath),
			fmt.Sprintf("-o=%s", htmlFilePath),
		); htmlRet != 0 {
			log.Printf("ERROR: Failed to output: %v", htmlFilePath)
			ret = 1
		}
	}
	return ret
}

func runAngularUnitTest(option testOption) int {
	popd := pushd("client")
	defer popd()

	args := []string{
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
		filepath.Join(option.LogDir, "server.stdout.log"),
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer serverStdout.Close()
	serverStderr, err := os.OpenFile(
		filepath.Join(option.LogDir, "server.stderr.log"),
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
	args := []string{
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
	client := &http.Client{
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

	e2eArgs := []string{
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
