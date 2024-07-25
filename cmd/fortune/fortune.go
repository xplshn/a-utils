// Copyright (c) 2024, xplshn [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/liamg/tml"
)

/* | fortune is a minimalistic implementation of the `fortune` program we all know
   | This version does not use an index file. Instead, it loads the entire fortune file into memory,
   | parses it, and randomly selects a fortune.
*/

// Version of the fortune program
const Version = "2.1"

// die prints a message to standard error and exits with a non-zero code
func die(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

// readFortuneFile reads and parses the fortune file into an array of fortunes
func readFortuneFile(fortuneFile string) ([]string, error) {
	content, err := ioutil.ReadFile(fortuneFile)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(content), "\n%\n"), nil
}

// findAndPrint selects and prints a random fortune from the given file
func findAndPrint(fortuneFile string) error {
	fortunes, err := readFortuneFile(fortuneFile)
	if err != nil {
		return err
	}
	rand.Seed(time.Now().UTC().UnixNano())
	fortune := fortunes[rand.Intn(len(fortunes))]
	fmt.Println(tml.Sprintf(fortune))
	return nil
}

// getRandomFortuneFile selects a random fortune file from the directories in the specified path
func getRandomFortuneFile(fortunePath string) (string, error) {
	paths := strings.Split(fortunePath, ":")
	var files []string

	for _, dir := range paths {
		dirFiles, err := ioutil.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, file := range dirFiles {
			if !file.IsDir() {
				files = append(files, filepath.Join(dir, file.Name()))
			}
		}
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no valid fortune files found in the specified paths")
	}

	rand.Seed(time.Now().UTC().UnixNano())
	return files[rand.Intn(len(files))], nil
}

// parseArgs parses the command line arguments and environment variables
func parseArgs() (string, error) {
	var fortuneFile string
	var fortunePath string

	flag.StringVar(&fortuneFile, "file", "", "Path to the fortune file")
	flag.StringVar(&fortunePath, "path", "", "Colon-separated list of directories containing fortune files")
	DisplayVersion := flag.Bool("version", false, "Display the version of this implementation")
	flag.Usage = func() {
		p := `
 Copyright (c) 2024, xplshn [3BSD]
 For more details refer to https://github.com/xplshn/a-utils

  Description
    Provide a quote from a "cookie file"
  Synopsis:
    fortune <--file|--path [file||directory]|--version>
  Options:
    --file: will read the provided file and echo a quote from it
    --list: will randomly select a file from within this directory and immeadiately return a quote from it
    --version: will display the version of this program
`
		fmt.Println(p)
	}
	flag.Parse()

	if *DisplayVersion {
		fmt.Println("a-utils's Fortune implementation is currently at version:", Version)
		os.Exit(5)
	}

	if fortuneFile != "" && fortunePath != "" {
		return "", fmt.Errorf("cannot use both -file and -path options at the same time")
	}

	if fortuneFile == "" && fortunePath == "" {
		fortuneFile = os.Getenv("FORTUNE_FILE")
		fortunePath = os.Getenv("FORTUNE_PATH")
	}

	if fortuneFile != "" {
		return fortuneFile, nil
	}

	if fortunePath != "" {
		return getRandomFortuneFile(fortunePath)
	}

	return "", fmt.Errorf("no fortune file specified and no FORTUNE_FILE or FORTUNE_PATH environment variable set")
}

// main is the entry point of the program
func main() {
	fortuneFile, err := parseArgs()
	if err != nil {
		die(err.Error())
	}

	err = findAndPrint(fortuneFile)
	if err != nil {
		die(err.Error())
	}
}
