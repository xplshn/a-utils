// Copyright 2012-2017 the u-root Authors. All rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// cat concatenates files and prints them to stdout.
//
// Synopsis:
//
//	cat [-u] [FILES]...
//
// Description:
//
//	If no files are specified, read from stdin.
//
// Options:
//
//	-u: Stub flag
//	-x: Enable syntax highlighting
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/u-root/u-root/pkg/uflag"
)

var (
	syntaxHighlighting bool
	uflagFile          = "/etc/cat.flags"
)

func init() {
	flag.BoolVar(&syntaxHighlighting, "x", false, "enable syntax highlighting")
	flag.Parse()
	if err := parseUFlagFile(uflagFile); err != nil {
		//		fmt.Fprintln(os.Stderr, "failed to parse uflag file:", err)
	}
}

func main() {
	args := flag.Args()

	if isOutputToTerminal() && !syntaxHighlighting {
		syntaxHighlighting = false
	}

	if err := run(os.Stdin, os.Stdout, syntaxHighlighting, args...); err != nil {
		fmt.Fprintln(os.Stderr, "cat failed:", err)
		os.Exit(1)
	}
}

func parseUFlagFile(filename string) error {
	contents, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	uflagArgs := uflag.FileToArgv(string(contents))
	uflagFlagSet := flag.NewFlagSet("uflag", flag.ContinueOnError)
	uflagFlagSet.BoolVar(&syntaxHighlighting, "x", syntaxHighlighting, "enable syntax highlighting")
	return uflagFlagSet.Parse(uflagArgs)
}

func isOutputToTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "could not stat stdout:", err)
		os.Exit(1)
	}
	return fi.Mode()&os.ModeCharDevice == 0
}

func run(stdin io.Reader, stdout io.Writer, syntaxHighlighting bool, args ...string) error {
	if len(args) == 0 {
		return processInput(stdin, stdout, "stdin", syntaxHighlighting)
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
		if err := processInput(reader, stdout, file, syntaxHighlighting); err != nil {
			return err
		}
	}
	return nil
}

func processInput(reader io.Reader, writer io.Writer, fileName string, syntaxHighlighting bool) error {
	if syntaxHighlighting {
		return highlightCat(reader, writer, fileName)
	}
	return cat(reader, writer)
}

func cat(reader io.Reader, writer io.Writer) error {
	_, err := io.Copy(writer, reader)
	return err
}

func highlightCat(reader io.Reader, writer io.Writer, fileName string) error {
	// Detect the lexer based on the fileName or use the fallback
	lexer := lexers.Match(fileName)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	// Get the style or use the fallback
	style := styles.Get("swapoff")
	if style == nil {
		style = styles.Get(os.Getenv("A_SYHX_COLOR_SCHEME"))
	}
	if style == nil {
		style = styles.Fallback
	}

	// Get the formatter or use the fallback
	formatterName := os.Getenv("A_SYHX_FORMAT")
	if formatterName == "" {
		formatterName = "terminal16" // Use a default formatter name
	}
	formatter := formatters.Get(formatterName)
	if formatter == nil {
		formatter = formatters.Fallback
	}

	// Read all content from the reader
	contents, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Tokenize the entire content
	iterator, err := lexer.Tokenise(nil, string(contents))
	if err != nil {
		return err
	}

	// Format the tokens and write to the writer
	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return err
	}

	// Write the formatted output
	_, err = io.Copy(writer, &buf)
	return err
}
