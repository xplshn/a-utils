// Copyright (c) 2024 xplshn                            [3BSD]
// For more details refer to https://github.com/xplshn/a-utils

// +build linux

package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

// getCPUInfo retrieves the CPU architecture and bit size using regular expressions.
func getCPUInfo() (string, string) {
	instructionSet := runtime.GOARCH
	bits := "unknown"

	// Define regular expressions for detecting bitness
	re64 := regexp.MustCompile(`(^64|64$)`)
	re32 := regexp.MustCompile(`(^32|32$|x86)`)

	if re64.MatchString(instructionSet) {
		bits = "64"
	} else if re32.MatchString(instructionSet) {
		bits = "32"
	}

	return instructionSet, bits
}

// getCPUFlags retrieves CPU flags using shell commands.
func getCPUFlags() (string, error) {
	cmd := exec.Command("sh", "-c", "cat /proc/cpuinfo | grep 'flags' | cut -f 2 -d ':' | awk '{$1=$1}1' | uniq")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get CPU flags: %v", err)
	}
	return string(output), nil
}

// printBits prints the CPU bits.
func printBits(bits string) {
	fmt.Println(bits)
}

// printInstSet prints the CPU instruction set.
func printInstSet(instructionSet string) {
	fmt.Println(instructionSet)
}

// printFlags prints the CPU flags.
func printFlags() {
	flags, err := getCPUFlags()
	if err != nil {
		fmt.Println("Error retrieving CPU flags:", err)
		return
	}
	fmt.Printf("CPU Flags: %s\n", flags)
}

// printHelp prints the help message.
func printHelp(option string) {
	fmt.Printf("isainfo: illegal option -%s\nusage: isainfo [-b | -k | -n | -x ]\n", option)
}

// main function to process command-line arguments and execute corresponding actions.
func main() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "isainfo",
		Synopsis:    "<|-b|-k|-n|-x>",
		Description: "Prints CPU architecture and flags.",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		os.Exit(1)
	}

	bitsFlag := flag.Bool("b", false, "Print the number of bits")
	instSetFlag := flag.Bool("k", false, "Print the instruction set")
	flagsFlag := flag.Bool("x", false, "Print the CPU flags")

	flag.Usage = func() { fmt.Print(helpPage) }

	flag.Parse()
	args := flag.Args()

	instructionSet, bits := getCPUInfo()

	if len(args) > 0 {
		printHelp(args[0])
		return
	}

	if *bitsFlag {
		printBits(bits)
	} else if *instSetFlag {
		printInstSet(instructionSet)
	} else if *flagsFlag {
		printFlags()
	} else {
		printInstSet(instructionSet)
	}
}
