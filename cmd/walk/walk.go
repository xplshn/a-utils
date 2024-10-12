// walk traverses directory trees, printing visited files // heavily modified version of ![Torgo's walk.go](https://github.com/as/torgo/blob/master/walk/walk.go)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const Prefix = "walk: "

type directory struct {
	Name  string
	Files []os.DirEntry
	Level int64
}

var (
	MaxConcurrency     = 1024 * 16
	goroutineSemaphore = make(chan struct{}, MaxConcurrency)
)

var run struct {
	traverse   func(string, func(string), int64)
	conditions []func(string) bool
	print      func(string)
}

var (
	visitedMap  = make(map[string]bool)
	rwlock      sync.RWMutex
	visitedFunc = markVisited
)

type FileInfo struct {
	Path       string
	IsDir      bool
	IsSymlink  bool
}

func isdirectory(path string) bool {
	FileInfo, err := gatherFileInfo(path)
	if err != nil {
		printError(err)
		return false
	}
	return FileInfo.IsDir
}

func isNotdirectory(path string) bool {
	return !isdirectory(path)
}

// isHidden checks if the given path has any hidden directories or if the filename itself is hidden.
func isHidden(path string) bool {
	// Split the path into components
	components := strings.Split(path, "/")

	// Extract filename and check if it is hidden
	filename := components[len(components)-1] // Last component is the filename
	if strings.HasPrefix(filename, ".") {
		return true // Filename is hidden
	}

	// Check if any component (directory) is hidden
	for _, component := range components {
		if strings.HasPrefix(component, ".") {
			return true // Found a hidden directory
		}
	}

	return false // No hidden directories or filename found
}

func isNotHidden(path string) bool {
	return !isHidden(path)
}

// markVisited marks a file as visited
func markVisited(path string) (alreadyVisited bool) {
	rwlock.RLock()
	_, alreadyVisited = visitedMap[path]
	rwlock.RUnlock()

	if !alreadyVisited {
		rwlock.Lock()
		visitedMap[path] = true
		rwlock.Unlock()
	}
	return
}

// printFile prints the file path if it meets all conditions
func printFile(path string) {
	for _, condition := range run.conditions {
		if !condition(path) {
			return
		}
	}
	fmt.Println(path)
}

// printAbsoluteFile prints the absolute path of the file if it meets all conditions
func printAbsoluteFile(path string) bool {
	var err error
	if !filepath.IsAbs(path) {
		if path, err = filepath.Abs(path); err != nil {
			printError(err)
			return false
		}
	}
	printFile(path)
	return true
}

func main() {
	// Options
	printRelative := flag.Bool("r", false, "Print relative paths")
	traversalLimit := flag.Int64("t", 1024*1024*1024, "Traversal depth limit")
	// Conditions
	printDirectories := flag.Bool("d", false, "Print directories only")
	printFiles := flag.Bool("f", false, "Print files only")
	showAll := flag.Bool("a", false, `Show paths that contain a directory or file prepended with '.'`)

	cmdInfo := &ccmd.CmdInfo{
		Name:        "walk",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "traverse a list of targets (directories or files)",
		Synopsis:    "<|-a|-nr|-t [INT]|> <|-d|-f|-x|-A|-a|> [target ...]", // FIX -x
	}
	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}
	flag.Usage = func() {
		fmt.Print(helpPage)
	}
	flag.Parse()

	if *printDirectories && *printFiles {
		printError("bad args: -dirs-only and -files-only cannot both be true")
		os.Exit(1)
	}

	run.traverse = depthFirstTraversal
	if !*showAll {
		run.conditions = append(run.conditions, isNotHidden)
	}
	if *printDirectories {
		run.conditions = append(run.conditions, isdirectory)
	}
	if *printFiles {
		run.conditions = append(run.conditions, isNotdirectory)
	}

	paths := flag.Args()
	// Default to current directory if no paths are specified
	if len(paths) == 0 {
		paths = []string{"."}
	}

	if *printRelative {
		run.print = func(path string) {
			for _, condition := range run.conditions {
				if !condition(path) {
					return
				}
			}
			fmt.Println(path)
		}
	} else {
		run.print = func(path string) {
			printAbsoluteFile(path)
		}
	}

	var wg sync.WaitGroup
	for _, target := range paths {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if target != "-" {
				run.traverse(target, run.print, *traversalLimit)
			} else {
				in := bufio.NewScanner(os.Stdin)
				for in.Scan() {
					run.traverse(in.Text(), run.print, *traversalLimit)
				}
			}
		}(target)
	}
	wg.Wait()
}

// depthFirstTraversal performs a depth-first traversal of the directory tree
func depthFirstTraversal(root string, fn func(string), maxDepth int64) {
	listChannel := make(chan string, MaxConcurrency)
	var wg sync.WaitGroup
	wg.Add(1)
	go depthFirstTraversalHelper(root, &wg, 0, listChannel, maxDepth)
	go func() {
		wg.Wait()
		close(listChannel)
	}()
	for path := range listChannel {
		run.print(path)
	}
}

// depthFirstTraversalHelper assists in depth-first traversal, enforcing depth limit
func depthFirstTraversalHelper(path string, wg *sync.WaitGroup, depth int64, listch chan<- string, maxDepth int64) {
	defer wg.Done()
	dir := getdirectory(path, depth)
	if dir == nil || depth > maxDepth {
		return
	}
	for _, file := range dir.Files {
		childPath := filepath.Join(dir.Name, file.Name())
		if visitedFunc(childPath) {
			continue
		}
		if file.IsDir() && depth < maxDepth {
			wg.Add(1)
			go depthFirstTraversalHelper(childPath, wg, depth+1, listch, maxDepth)
		}
		listch <- childPath
	}
}

// getdirectory reads the directory and returns its contents
func getdirectory(path string, level int64) *directory {
	files, err := os.ReadDir(path)
	if err != nil || files == nil {
		printError(err)
		return nil
	}
	return &directory{Name: path, Files: files, Level: level}
}

// println prints a message with the prefix
func println(v ...interface{}) {
	fmt.Print(Prefix)
	fmt.Println(v...)
}

// printError prints an error message with the prefix
func printError(v ...interface{}) {
	fmt.Fprint(os.Stderr, Prefix)
	fmt.Fprintln(os.Stderr, v...)
}

// timespecToTime converts syscall.Timespec to time.Time
func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}
