// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "envsubst",
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Substitutes environment variables in shell format strings",
		Synopsis:    "[OPTION] [SHELL-FORMAT]",
		CustomFields: map[string]interface{}{
			"1_Behavior": "By default, reads stdin and substitutes environment variables.",
			"2_Examples":  `  \$ echo 'Hello $USER' | envsubst`,
		},
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	var variablesFlag bool

	flag.BoolVar(&variablesFlag, "v", false, "output the variables occurring in SHELL-FORMAT")
	flag.BoolVar(&variablesFlag, "variables", false, "output the variables occurring in SHELL-FORMAT")
	flag.Parse()

	args := flag.Args()

	// Check if no arguments and no stdin
	stat, _ := os.Stdin.Stat()
	if len(args) == 0 && (stat.Mode()&os.ModeCharDevice != 0) {
		// If there are no args and no stdin, print the help page
		flag.Usage()
		return
	}

	// If variables flag is set, output the variables found in SHELL-FORMAT
	if variablesFlag {
		if len(args) > 0 {
			fmt.Println(extractVariables(args[0]))
		}
		return
	}

	// Read from stdin if no shell format is provided
	input := ""
	if len(args) == 0 {
		inputBytes, err := os.ReadFile("/dev/stdin")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
			os.Exit(1)
		}
		input = string(inputBytes)
	} else {
		input = args[0]
	}

	// Substitute environment variables
	output := substituteEnvVariables(input)
	fmt.Print(output)
}

// Extracts environment variable names from a shell format string
func extractVariables(shellFormat string) string {
	re := regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?`)
	matches := re.FindAllStringSubmatch(shellFormat, -1)

	var vars []string
	for _, match := range matches {
		vars = append(vars, match[1])
	}
	return strings.Join(vars, "\n")
}

// Substitutes environment variables in the input string
func substituteEnvVariables(input string) string {
	re := regexp.MustCompile(`\$\{?([A-Za-z_][A-Za-z0-9_]*)\}?`)
	return re.ReplaceAllStringFunc(input, func(v string) string {
		varName := strings.Trim(v, "${}")
		return os.Getenv(varName)
	})
}
