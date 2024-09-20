// Copyright 2016 "as".

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	Prefix = "fi: "
)

func main() {
	printSize := flag.Bool("s", false, "Print file size")
	printCumulative := flag.Bool("c", false, "Print cumulative statistics")
	printModTime := flag.Bool("m", false, "Print modified time")
	printDuration := flag.Bool("d", false, "Print duration since last modified")
	printPermissions := flag.Bool("p", false, "Print file permissions")
	verbose := flag.Bool("v", false, "Print verbose error messages")

	cmdInfo := &ccmd.CmdInfo{
		Name:        "fi",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Prints file information for files read from stdin",
		Synopsis:    "fi [-s -c -m -d -p -v]",
		CustomFields: map[string]interface{}{
			"1_Examples": `Print file sizes and cumulative total:
  \$ walk -f mink/ | fi -s -c`,
		},
	}
	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}
	// Set usage to print the CCMD-generated help page
	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	// Parse flags
	flag.Parse()

	scanner := bufio.NewScanner(os.Stdin)
	var printFuncs []func(string, os.FileInfo) string

	// Build the list of printing functions based on the flags
	printFuncs = append(printFuncs, func(s string, fi os.FileInfo) string {
		return fmt.Sprintf("%s\n", s)
	})
	if *printSize {
		printFuncs = append(printFuncs, func(s string, fi os.FileInfo) string {
			return fmt.Sprintf("%d\t", fi.Size())
		})
	}
	if *printDuration {
		printFuncs = append(printFuncs, func(s string, fi os.FileInfo) string {
			return fmt.Sprintf("%s\t", time.Since(fi.ModTime()).Truncate(time.Second))
		})
	}
	if *printModTime {
		printFuncs = append(printFuncs, func(s string, fi os.FileInfo) string {
			return fmt.Sprintf("%s\t", fi.ModTime().Format("2006.01.02 15:04:05"))
		})
	}
	if *printPermissions {
		printFuncs = append(printFuncs, func(s string, fi os.FileInfo) string {
			return fmt.Sprintf("%s\t", fi.Mode())
		})
	}

	totalSize := int64(0)

	// Process each file from stdin
	for scanner.Scan() {
		fileName := scanner.Text()
		fileInfo, err := os.Stat(fileName)
		if err != nil {
			if *verbose {
				printError(err)
			}
			continue
		}

		output := ""
		for i := len(printFuncs) - 1; i >= 0; i-- {
			output += printFuncs[i](fileName, fileInfo)
		}
		fmt.Print(output)
		totalSize += fileInfo.Size()
	}

	// Print cumulative total if the flag is set
	if *printSize && *printCumulative {
		fmt.Printf("%d total\n", totalSize)
	}
}

func printError(v ...interface{}) {
	fmt.Fprint(os.Stderr, Prefix)
	fmt.Fprintln(os.Stderr, v...)
}
