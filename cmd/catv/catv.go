// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"

	"a-utils/pkg/ccmd"
)

// removeAnsiSequences removes ANSI escape sequences from the given content.
func removeAnsiSequences(content string) string {
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9]*(?:;[0-9]*)*[a-zA-Z]`)
	return ansiEscape.ReplaceAllString(content, "")
}

// makeNonPrintableVisible converts non-printable characters to a visible format.
func makeNonPrintableVisible(content string, showTabs bool, showEndLines bool) string {
	visibleContent := ""
	for i, char := range content {
		switch char {
		case '\n':
			visibleContent += "\n"
			if showEndLines && i != len(content)-1 {
				visibleContent += "$"
			}
		case '\t':
			if showTabs {
				visibleContent += "^I"
			} else {
				visibleContent += "\t"
			}
		default:
			if char < ' ' || char == '\x7f' {
				visibleContent += fmt.Sprintf("^%c", char+'@')
			} else {
				visibleContent += string(char)
			}
		}
	}
	if showEndLines && len(content) > 0 && content[len(content)-1] == '\n' {
		visibleContent += "$"
	}
	return visibleContent
}

// processFile reads the file content, processes it according to the flags, and writes the result to the output.
func processFile(reader io.Reader, writer io.Writer, removeAnsiFlag bool, showNonPrintable bool, showTabs bool, showEndLines bool) error {
	content, err := io.ReadAll(reader)
	if err != nil {
		return fmt.Errorf("error reading input: %v", err)
	}

	var finalContent string
	if removeAnsiFlag {
		finalContent = removeAnsiSequences(string(content))
	} else if showNonPrintable {
		finalContent = makeNonPrintableVisible(string(content), showTabs, showEndLines)
	} else {
		finalContent = string(content)
	}

	_, err = writer.Write([]byte(finalContent))
	return err
}

// run processes each file provided in args or reads from stdin if no files are specified.
func run(stdin io.Reader, stdout io.Writer, args []string, removeAnsiFlag bool, showNonPrintable bool, showTabs bool, showEndLines bool) error {
	if len(args) == 0 {
		return processFile(stdin, stdout, removeAnsiFlag, showNonPrintable, showTabs, showEndLines)
	}

	for _, file := range args {
		var reader io.Reader
		if file == "-" {
			reader = stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", file, err)
			}
			defer f.Close()
			reader = f
		}
		if err := processFile(reader, stdout, removeAnsiFlag, showNonPrintable, showTabs, showEndLines); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Name:        "catv",
		Synopsis:    "<|-v|-t|-e|-r|-A|> [FILE/s]",
		Description: "Provides a non-harmful way to make non-printable characters visible from the specified files",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		return
	}

	removeAnsiFlag := flag.Bool("r", false, "Remove ANSI escape sequences")
	showNonPrintable := flag.Bool("v", false, "Show non-printing characters as ^x or M-x")
	showTabs := flag.Bool("t", false, "Show tabs as ^I")
	showEndLines := flag.Bool("e", false, "Show end of lines with $")
	showAll := flag.Bool("A", false, "Same as -vte")

	flag.Usage = func() { fmt.Print(helpPage) }

	flag.Parse()

	args := flag.Args()

	// If no flags are used, set showNonPrintable to true
	if !*removeAnsiFlag && !*showTabs && !*showEndLines && !*showAll {
		*showNonPrintable = true
	}

	if *showAll {
		*showNonPrintable = true
		*showTabs = true
		*showEndLines = true
	}

	if err := run(os.Stdin, os.Stdout, args, *removeAnsiFlag, *showNonPrintable, *showTabs, *showEndLines); err != nil {
		fmt.Fprintln(os.Stderr, "catv failed:", err)
		os.Exit(1)
	}
}
