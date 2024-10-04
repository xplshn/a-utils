// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

// displayCPUInfo prints CPU info based on the selected options.
func displayCPUInfo(showBits, showInstSet, showFlags, showVendor, showCores, showMhz, showISAVersion bool) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error retrieving CPU info: %v\n", err)
		os.Exit(1)
	}

	// Use information from the first CPU (since all cores are typically the same)
	info := cpuInfo[0]

	if showBits {
		bits := "unknown"
		if runtime.GOARCH == "amd64" || runtime.GOARCH == "arm64" {
			bits = "64"
		} else if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
			bits = "32"
		}
		fmt.Printf("Bits: %s\n", bits)
	}

	if showInstSet {
		fmt.Printf("Instruction Set: %s\n", info.ModelName)
	}

	if showFlags {
		fmt.Printf("Flags: %s\n", strings.Join(info.Flags, " "))
	}

	if showVendor {
		fmt.Printf("Vendor: %s\n", info.VendorID)
	}

	if showCores {
		// Print total number of cores across all CPUs
		totalCores, err := cpu.Counts(true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error retrieving CPU cores count: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Cores: %d\n", totalCores)
	}

	if showMhz {
		fmt.Printf("MHz: %.2f\n", info.Mhz)
	}

	if showISAVersion {
		isaVersion := determineISAVersion(info.Flags)
		fmt.Printf("ISA Version: %s\n", isaVersion)
	}
}

// determineISAVersion checks the CPU flags and returns the ISA version.
func determineISAVersion(flags []string) string {
	flagSet := make(map[string]bool)
	for _, flag := range flags {
		flagSet[flag] = true
	}

	if hasFlags(flagSet, "avx512f", "avx512bw", "avx512cd", "avx512dq", "avx512vl") {
		return "x86-64-v4"
	}
	if hasFlags(flagSet, "avx", "avx2", "bmi1", "bmi2", "f16c", "fma", "abm", "movbe", "xsave") {
		return "x86-64-v3"
	}
	if hasFlags(flagSet, "cx16", "lahf_lm", "popcnt", "sse4_1", "sse4_2", "ssse3") {
		return "x86-64-v2"
	}
	if hasFlags(flagSet, "lm", "cmov", "cx8", "fpu", "fxsr", "mmx", "syscall", "sse2") {
		return "x86-64-v1"
	}
	return "unknown"
}

// hasFlags checks if all the given flags are present in the flag set.
func hasFlags(flagSet map[string]bool, flags ...string) bool {
	for _, flag := range flags {
		if !flagSet[flag] {
			return false
		}
	}
	return true
}

// printHelp prints a custom help message for invalid options.
func printHelp(option string) {
	fmt.Printf("isainfo: illegal option -%s\nusage: isainfo [-b|-k|-x|-v|-c|-m|-i]\n", option)
}

// main function processes command-line arguments and displays CPU information based on user options.
func main() {
	bitsFlag := flag.Bool("b", false, "Print the number of bits (32 or 64)")
	instSetFlag := flag.Bool("k", false, "Print the instruction set (model name)")
	flagsFlag := flag.Bool("x", false, "Print the CPU flags")
	vendorFlag := flag.Bool("v", false, "Print the CPU vendor ID")
	coresFlag := flag.Bool("c", false, "Print the number of CPU cores")
	mhzFlag := flag.Bool("m", false, "Print the CPU clock speed (MHz)")
	isaVersionFlag := flag.Bool("i", false, "Print the ISA version")

	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "isainfo",
		Synopsis:    "<|-b|-k|-x|-v|-c|-m|-iv|>",
		Description: "Prints detailed CPU architecture and flags.",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		os.Exit(1)
	}

	flag.Usage = func() { fmt.Print(helpPage) }
	flag.Parse()

	// If no flags are provided, print usage
	if !*bitsFlag && !*instSetFlag && !*flagsFlag && !*vendorFlag && !*coresFlag && !*mhzFlag && !*isaVersionFlag {
		flag.Usage()
		os.Exit(0)
	}

	// Display the CPU information based on flags
	displayCPUInfo(*bitsFlag, *instSetFlag, *flagsFlag, *vendorFlag, *coresFlag, *mhzFlag, *isaVersionFlag)
}
