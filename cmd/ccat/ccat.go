// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"a-utils/pkg/ccmd"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "ccat",
		Authors:     []string{"xplshn"},
		Description: "Concatenates files and prints them to stdout with Syntax Highlighting",
		Synopsis:    "[FILE/s]",
		Behavior:    "If no files are specified, read from stdin.",
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
	if err := run(os.Stdin, os.Stdout, args...); err != nil {
		fmt.Fprintln(os.Stderr, "cat failed:", err)
		os.Exit(1)
	}
}

func run(stdin io.Reader, stdout io.Writer, args ...string) error {
	if len(args) == 0 {
		return processInput(stdin, stdout, "stdin")
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
		if err := processInput(reader, stdout, file); err != nil {
			return err
		}
	}
	return nil
}

func processInput(reader io.Reader, writer io.Writer, fileName string) error {
	return highlightCat(reader, writer, fileName)
}

func highlightCat(reader io.Reader, writer io.Writer, fileName string) error {
	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	style := styles.Get("swapoff")
	if style == nil {
		style = styles.Get(os.Getenv("A_SYHX_COLOR_SCHEME"))
	}
	if style == nil {
		style = styles.Fallback
	}

	formatterName := os.Getenv("A_SYHX_FORMAT")
	if formatterName == "" {
		formatterName = "terminal16"
	}
	formatter := formatters.Get(formatterName)
	if formatter == nil {
		formatter = formatters.Fallback
	}

	contents, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	iterator, err := lexer.Tokenise(nil, string(contents))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return err
	}

	_, err = io.Copy(writer, &buf)
	return err
}
