package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/shirou/gopsutil/v4/process"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "importenv",
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Run a given command, with the ENV of a PID",
		Synopsis:    "[[PID] [COMMAND]]",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}
	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()

	args := flag.Args()

	if len(args) < 2 {
		flag.Usage()
		return
	}

	pid, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid PID:", err)
		return
	}

	command := args[1]
	args = args[2:]

	proc, err := process.NewProcess(int32(pid))
	if err != nil {
		fmt.Println("Failed to get process:", err)
		return
	}

	env, err := proc.Environ()
	if err != nil {
		fmt.Println("Failed to get environment variables:", err)
		return
	}

	cmd := exec.Command(command, args...)
	cmd.Env = env

	output, _ := cmd.CombinedOutput()

	fmt.Print(string(output))
}
