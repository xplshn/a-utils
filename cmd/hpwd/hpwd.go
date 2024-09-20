// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	cmdInfo := ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "hpwd",
		Usage:       "<|-h>",
		Description: "Stylized `pwd` command",
		CustomFields: map[string]interface{}{
			"1_Behavior": "If in the home directory, display '~' or the value of COOLHOME if set. If inside the home directory but not in the home itself, display the relative path prefixed with COOLHOME_DEPTH if set. Otherwise, display the full path.",
		},
		Options: []string{"-h", "--help"},
	}

	// Check for help flags manually
	for _, arg := range os.Args[1:] {
		if arg == "-h" || arg == "--help" {
			helpPage, err := cmdInfo.GenerateHelpPage()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error generating help page:", err)
				os.Exit(1)
			}
			fmt.Print(helpPage)
			os.Exit(0)
		}
	}

	// Main logic of the command
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting current directory:", err)
		return
	}

	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	coolHome := os.Getenv("COOLHOME")
	coolHomeDepth := os.Getenv("COOLHOME_DEPTH")

	if pwd == home {
		if coolHome != "" {
			fmt.Print(coolHome)
		} else {
			fmt.Print("~") // Display '~' if in home directory
		}
	} else if strings.HasPrefix(pwd, home+"/") { // Adjusted condition
		// Calculate the relative path from home
		relativePath := pwd[len(home)+1:]
		if coolHomeDepth != "" {
			fmt.Printf("%s/%s ", coolHomeDepth, relativePath)
		} else {
			fmt.Printf("~/%s ", relativePath) // Display '~/...' for directories inside home if COOLHOME_DEPTH is not set
		}
	} else {
		fmt.Print(pwd + " ") // Display the full path if outside home
	}
}
