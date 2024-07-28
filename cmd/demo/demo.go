// Copyright (c) 2024, xplshn [3BSD]
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
		Authors:      []string{"John Doe", "Jane Smith"},
		Name:         "example",
		Synopsis:     "command [options]",
		Description:  "This is an example command line tool to demonstrate the ccmd library.",
		Behavior:     "This tool demonstrates basic usage of the ccmd library and formatting.",
		Notes:        "Make sure to use the correct flags and options as shown below.",
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
	paragraph := `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	header := "long text"
	fmt.Println(ccmd.CFormatCenter(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
	paragraph = `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	header = "long text, left-aligned (will rob you and tax you)"
	fmt.Println(ccmd.CFormatLeft(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
}
