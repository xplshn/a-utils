// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"a-utils/pkg/ccmd"
)

func main() {
	// Create an instance of CmdInfo
	cmdInfo := &ccmd.CmdInfo{
		Authors:      []string{"John Doe", "Jane Smith", "Joe Momma"},
		Name:         "demo",
		Synopsis:     "demo [options]",
		Description:  "This is an example command line tool to demonstrate the ccmd (consistent command line) library.",
		Behavior:     "This tool demonstrates basic usage of the ccmd library and formatting.",
		Notes:        "we good",
		ExcludeFlags: make(map[string]bool),
		Since:        1999,
	}

	// Define some flags
	cmdInfo.DefineFlag("verbose", false, "Enable verbose output", false)
	cmdInfo.DefineFlag("count", 1, "Number of times to repeat", false)
	cmdInfo.DefineFlag("name", "user", "Name to greet", false)
	cmdInfo.DefineFlag("duration", time.Second, "Duration to wait", false)

	// Generate the help page
	helpText, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpText)
	}

	// Parse command line flags
	flag.Parse()

	// Print the centered help page
	fmt.Println(ccmd.FormatCenter("Centered text 1"))
	header := "long text (centered)"
	paragraph := `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	fmt.Println(ccmd.CFormatCenter(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
	header = "long text, left-aligned (will rob you and tax you)"
	paragraph = `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	fmt.Println(ccmd.CFormatLeft(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
	header = "long text, right-aligned (will rob you and tax you, its not a replacement for a libertarian, but says it is)"
	paragraph = `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	fmt.Println(ccmd.CFormatRight(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
}
