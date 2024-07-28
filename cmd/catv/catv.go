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

func makeVisible(content string) string {
	// Replace non-printable characters with their visible representations
	visibleContent := ""
	for _, char := range content {
		switch char {
		case '\n':
			visibleContent += "\n"
		case '\t':
			visibleContent += "\t"
		default:
			if char < ' ' || char == '\x7f' {
				visibleContent += fmt.Sprintf("^%c", char+'@')
			} else {
				visibleContent += string(char)
			}
		}
	}
	return visibleContent
}

func printFileWithOptions(filename string, removeAnsiFlag bool, showAnsiFlag bool) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filename, err)
		return
	}

	var finalContent string
	if removeAnsiFlag {
		finalContent = removeAnsi(string(content))
	} else if showAnsiFlag {
		finalContent = string(content)
	} else {
		finalContent = makeVisible(string(content))
	}

	fmt.Println(finalContent)
}

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Name:        "catv",
		Synopsis:    "<|-r|-s|> [FILE/s]",
		Description: "Provides a non-harmful way to remove ANSI escape sequences and make non-printable characters visible from the specified files.",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		return
	}

	removeAnsiFlag := flag.Bool("r", false, "Remove ANSI escape sequences")
	showAnsiFlag := flag.Bool("s", false, "Show ANSI escape sequences without interpreting them")

	flag.Usage = func() { fmt.Print(helpPage) }
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprint(os.Stderr, helpPage)
		fmt.Fprintln(os.Stderr, "error: no valid arguments were provided.")
		return
	}

	for _, filename := range flag.Args() {
		printFileWithOptions(filename, *removeAnsiFlag, *showAnsiFlag)
	}
}
