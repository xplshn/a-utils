// Copyright 2009-2018 the u-root Authors               [3BSD]
// Copyright 2024 xplshn
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/u-root/u-root/pkg/ldd"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	// Setup CCMD library for help and command information
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"u-root", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "lddfiles",
		Synopsis:    "[FILE/s]",
		Description: "Display all dynamic dependencies of the provided files using the  u-root ldd package.",
	}

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating help page: %v\n", err)
		return
	}

	flag.Usage = func() { fmt.Print(helpPage) }

	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "No files provided.")
		flag.Usage()
		os.Exit(1)
	}

	if err := run(os.Stdout, args); err != nil {
		log.Fatal(err)
	}
}

func run(stdout io.Writer, args []string) error {
	l, err := ldd.FList(args...)
	if err != nil {
		return fmt.Errorf("ldd: %w", err)
	}

	for _, p := range args {
		a, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("ldd: %w", err)
		}
		l = append(l, a)
	}

	for _, dep := range l {
		fmt.Fprintf(stdout, "%s\n", dep)
	}

	return nil
}
