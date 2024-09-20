// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"strconv"
	"strings"
	"unicode"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

// Main function
func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}

	if len(args) > 4 {
		fmt.Fprintf(os.Stderr, "too many arguments\n")
		os.Exit(1)
	}

	result := performTest(args)
	if result {
		os.Exit(0)
	}
	os.Exit(1)
}

// Perform test based on arguments
func performTest(args []string) bool {
	switch len(args) {
	case 0:
		return false
	case 1:
		return isNonEmptyString(args[0])
	case 2:
		if args[0] == "!" {
			return !performTest(args[1:])
		}
		if fn, ok := fileTests[args[0]]; ok {
			return fn(args[1])
		}
		fmt.Fprintf(os.Stderr, "bad unary test %s\n", args[0])
		return false
	case 3:
		if fn, ok := binaryTests[args[1]]; ok {
			return fn(args[0], args[2])
		}
		if args[0] == "!" {
			return !performTest(args[1:])
		}
		fmt.Fprintf(os.Stderr, "bad binary test %s\n", args[1])
		return false
	case 4:
		if args[0] == "!" {
			return !performTest(args[1:])
		}
		fmt.Fprintf(os.Stderr, "too many arguments\n")
		return false
	default:
		return false
	}
}

// Print help message
func printHelp() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "test",
		Synopsis:    "[-bcdefghkLprSsuwx PATH] [-nz STRING] [-t FD] [X ?? Y]",
		Description: "Return true or false by performing tests. No arguments is false, one argument is true if not empty string.",
		CustomFields: map[string]interface{}{
			"Notes": `--- Tests with a single argument (after the option):
				PATH is/has:
				  -b  block device   -f  regular file   -p  fifo           -u  setuid bit
				  -c  char device    -g  setgid         -r  readable       -w  writable
				  -d  directory      -h  symlink        -S  socket         -x  executable
				  -e  exists         -L  symlink        -s  nonzero size   -k  sticky bit
				STRING is:
				  -n  nonzero size   -z  zero size
				FD (integer file descriptor) is:
				  -t  a TTY

				--- non-POSIX tests: '-k', '[[' '<' '>' '=~' ']]'

				--- Tests with one argument on each side of an operator:
				Two strings:
				  =  are identical   !=  differ         =~  string matches regex
				Alphabetical sort:
				  <  first is lower  >   first higher
				Two integers:
				  -eq  equal         -gt  first > second    -lt  first < second
				  -ne  not equal     -ge  first >= second   -le  first <= second
				Two files:
				  -ot  Older mtime   -nt  Newer mtime       -ef  same dev/inode

				--- Modify or combine tests:
				  ! EXPR     not (swap true/false)   EXPR -a EXPR    and (are both true)
				  ( EXPR )   evaluate this first     EXPR -o EXPR    or (is either true)`,
		},
	}
	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Printf("Error generating help page: %v\n", err)
		return
	}
	fmt.Print(helpPage)
}

// Here be dragons ↴

// intcmp compares two integer strings, handling negative and positive signs,
// and taking into account the order of magnitude for comparison.
func intcmp(a, b string) int {
	// Determine the sign of each integer
	asign := 1
	bsign := 1
	if a[0] == '-' {
		asign = -1
	}
	if b[0] == '-' {
		bsign = -1
	}

	// Move past the sign character if present
	if a[0] == '-' || a[0] == '+' {
		a = a[1:]
	}
	if b[0] == '-' || b[0] == '+' {
		b = b[1:]
	}

	// Check for non-integer input
	if len(a) == 0 || len(b) == 0 {
		goto noint
	}
	for _, ch := range a {
		if !unicode.IsDigit(ch) {
			goto noint
		}
	}
	for _, ch := range b {
		if !unicode.IsDigit(ch) {
			goto noint
		}
	}

	// Remove leading zeros
	a = strings.TrimLeft(a, "0")
	b = strings.TrimLeft(b, "0")
	if len(a) == 0 {
		asign = 0
	}
	if len(b) == 0 {
		bsign = 0
	}

	// Compare based on signs and lengths
	if asign != bsign {
		if asign < bsign {
			return -1
		}
		return 1
	}
	if len(a) != len(b) {
		if len(a) < len(b) {
			return asign * -1
		}
		return asign * 1
	}

	// Compare lexicographically if the lengths and signs are the same
	return asign * strings.Compare(a, b)

noint:
	fmt.Fprintf(os.Stderr, "expected integer operands\n")
	return 0
}

// isDigit checks if a string is a valid integer.
func isDigit(s string) bool {
	if len(s) == 0 {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

// mtimecmp compares modification times of two files.
func mtimecmp(info1, info2 os.FileInfo) int {
	if info1.ModTime().Before(info2.ModTime()) {
		return -1
	}
	if info1.ModTime().After(info2.ModTime()) {
		return 1
	}
	return 0
}

// File type checks
func isBlockDevice(path string) bool {
	return isType(path, os.ModeDevice|os.ModeCharDevice)
}
func isCharDevice(path string) bool {
	return isType(path, os.ModeDevice|os.ModeCharDevice)
}
func isDirectory(path string) bool {
	return isType(path, os.ModeDir)
}
func isRegularFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}
func isSetGID(path string) bool {
	return isMode(path, os.ModeSetgid)
}
func isSymlink(path string) bool {
	return isType(path, os.ModeSymlink)
}
func isSticky(path string) bool {
	return isMode(path, os.ModeSticky)
}
func isNamedPipe(path string) bool {
	return isType(path, os.ModeNamedPipe)
}
func isSocket(path string) bool {
	return isType(path, os.ModeSocket)
}
func hasSize(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Size() > 0
}
func isSetUID(path string) bool {
	return isMode(path, os.ModeSetuid)
}
func isNonEmptyString(s string) bool {
	return len(s) > 0
}
func isEmptyString(s string) bool {
	return len(s) == 0
}
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isReadable checks if the specified file path is readable.
func isReadable(path string) bool {
	file, err := os.Open(path)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

// isWritable checks if the specified file path is writable.
func isWritable(path string) bool {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	if err != nil {
		return false
	}
	file.Close()
	return true
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode()&0111 != 0
}

func isTerminal(fdStr string) bool {
	fd, err := strconv.Atoi(fdStr)
	if err != nil {
		return false
	}
	return isattyFd(fd)
}

// Binary tests
func stringsEqual(s1, s2 string) bool {
	return s1 == s2
}

func stringsNotEqual(s1, s2 string) bool {
	return s1 != s2
}

func integersEqual(s1, s2 string) bool {
	return intcmp(s1, s2) == 0
}

func integersNotEqual(s1, s2 string) bool {
	return intcmp(s1, s2) != 0
}

func integerGreaterThan(s1, s2 string) bool {
	return intcmp(s1, s2) > 0
}

func integerGreaterOrEqual(s1, s2 string) bool {
	return intcmp(s1, s2) >= 0
}

func integerLessThan(s1, s2 string) bool {
	return intcmp(s1, s2) < 0
}

func integerLessOrEqual(s1, s2 string) bool {
	return intcmp(s1, s2) <= 0
}

func filesEqual(s1, s2 string) bool {
	var stat1, stat2 unix.Stat_t

	// Get file descriptor for the first file
	fd1, err := unix.Open(s1, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", s1, err)
		os.Exit(1)
	}
	defer unix.Close(fd1)

	// Get file descriptor for the second file
	fd2, err := unix.Open(s2, unix.O_RDONLY, 0)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file %s: %v\n", s2, err)
		os.Exit(1)
	}
	defer unix.Close(fd2)

	// Retrieve the file stats
	if err := unix.Fstat(fd1, &stat1); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file stats for %s: %v\n", s1, err)
		os.Exit(1)
	}
	if err := unix.Fstat(fd2, &stat2); err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file stats for %s: %v\n", s2, err)
		os.Exit(1)
	}

	// Compare device ID and inode number
	return stat1.Dev == stat2.Dev && stat1.Ino == stat2.Ino
}

func fileOlderThan(s1, s2 string) bool {
	info1, err1 := os.Stat(s1)
	info2, err2 := os.Stat(s2)
	if err1 != nil || err2 != nil {
		return false
	}
	return mtimecmp(info1, info2) < 0
}

func fileNewerThan(s1, s2 string) bool {
	info1, err1 := os.Stat(s1)
	info2, err2 := os.Stat(s2)
	if err1 != nil || err2 != nil {
		return false
	}
	return mtimecmp(info1, info2) > 0
}

// Helper function
func isType(path string, mode os.FileMode) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode()&mode != 0
}

func isMode(path string, mode os.FileMode) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode()&mode != 0
}

// isattyFd checks if the given file descriptor is a terminal using the isatty system call.
func isattyFd(fd int) bool {
	_, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	return err == nil
}

// Test definitions
type fileTestFunc func(string) bool
type binaryTestFunc func(string, string) bool

var fileTests = map[string]fileTestFunc{
	"-b": isBlockDevice, "-c": isCharDevice, "-d": isDirectory, "-e": fileExists, "-f": isRegularFile,
	"-g": isSetGID, "-h": isSymlink, "-k": isSticky, "-L": isSymlink, "-n": isNonEmptyString,
	"-p": isNamedPipe, "-r": isReadable, "-S": isSocket, "-s": hasSize, "-t": isTerminal,
	"-u": isSetUID, "-w": isWritable, "-x": isExecutable, "-z": isEmptyString,
}

var binaryTests = map[string]binaryTestFunc{
	"=": stringsEqual, "!=": stringsNotEqual, "-eq": integersEqual, "-ne": integersNotEqual,
	"-gt": integerGreaterThan, "-ge": integerGreaterOrEqual, "-lt": integerLessThan, "-le": integerLessOrEqual,
	"-ot": fileOlderThan, "-nt": fileNewerThan, "-ef": filesEqual,
}
