// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	// Create an instance of CmdInfo
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"John Doe", "Jane Smith", "Joe Momma"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "demo",
		Synopsis:    "demo [options]",
		Description: "This is an example command line tool to demonstrate the ccmd (consistent command line) library.",
		CustomFields: map[string]interface{}{
			"1_Behavior": "This tool demonstrates basic usage of the ccmd library and formatting.",
			"2_Notes":    "We good",
		},
		Since: 1999,
	}

	// Define some flags using the flag package
	flag.Bool("verbose", false, "Enable verbose output")
	flag.Int("count", 1, "Number of times to repeat")
	flag.String("name", "user", "Name to greet")
	flag.Duration("duration", time.Second, "Duration to wait")

	flag.Usage = func() {
		helpPage, err := cmdInfo.GenerateHelpPage()
		if err != nil {
			fmt.Printf("Error generating help page: %v\n", err)
			return
		}
		fmt.Print(helpPage)
	}
	// Parse the flags
	flag.Parse()

	// Print the centered help page
	fmt.Println(ccmd.FormatCenter("Centered text 1"))
	header := "long text (centered)"
	paragraph := `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	fmt.Println(ccmd.CFormatCenter(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
	header = "long text, right-aligned (will rob you and tax you, its not a replacement for a libertarian, but says it is)"
	paragraph = `I am a very long paragraph, with more than one line and son
I expect you not read this, its just to test the header title alignment`
	fmt.Println(ccmd.CFormatRight(header, ccmd.RelativeTo(paragraph)))
	fmt.Println(paragraph)
	fmt.Println()
}
