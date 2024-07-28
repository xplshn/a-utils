// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"

	"a-utils/pkg/ccmd"
)

func removeAnsi(content string) string {
	// Regular expression to match ANSI escape sequences, avoiding ";" when not needed.
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9]*(?:;[0-9]*)*[a-zA-Z]`)
	return ansiEscape.ReplaceAllString(content, "")
}

func printFileWithoutAnsi(filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		return
	}

	cleanContent := removeAnsi(string(content))
	fmt.Println(cleanContent)
}

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn", "sweetbbak", "u-root"},
		Name:        "noansi",
		Synopsis:    "[FILE/s]",
		Description: "Remove ANSI escape sequences from the specified files.",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		return
	}

	flag.Usage = func() { fmt.Print(helpPage) }
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprint(os.Stderr, helpPage)
		fmt.Fprintln(os.Stderr, "error: no valid arguments were provided.")
		return
	}

	for _, filename := range flag.Args() {
		printFileWithoutAnsi(filename)
	}
}
