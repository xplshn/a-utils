// Copyright (c) 2024, xplshn [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
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
