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

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const Prefix = "walk: "

// Constants and global variables
var (
	MaxConcurrency     = 1024 * 16
	goroutineSemaphore = make(chan struct{}, MaxConcurrency)
)

var run struct {
	traverse   func(string, func(string), int64)
	conditions []func(string) bool
	print      func(string)
}

type Directory struct {
	Name  string
	Files []os.DirEntry
	Level int64
}

var (
	visitedMap  = make(map[string]bool)
	rwlock      sync.RWMutex
	visitedFunc = markVisited
)

// isDirectory checks if the given path is a directory
func isDirectory(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		printError(err)
		return false
	}
	return info.IsDir()
}

// isNotDirectory checks if the given path is not a directory
func isNotDirectory(path string) bool {
	return !isDirectory(path)
}

// isNotHidden checks if the given path is not a hidden file or directory
func isNotHidden(path string) bool {
	return !strings.HasPrefix(filepath.Base(path), ".")
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
	printAbsolute := flag.Bool("a", false, "Print absolute paths")
	noRelativePrefix := flag.Bool("nr", false, "Don't print ./ in relative paths")
	traversalLimit := flag.Int64("t", 1024*1024*1024, "Traversal depth limit")
	// Conditions
	printDirectories := flag.Bool("d", false, "Print directories only")
	printFiles := flag.Bool("f", false, "Print files only")
	hideHidden := flag.Bool("x", false, "Hide hidden files")

	cmdInfo := &ccmd.CmdInfo{
		Name:        "walk",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "traverse a list of targets (directories or files)",
		Synopsis:    "<|-a|-nr|-t [INT]|> <|-d|-f|-x|> [target ...]", // FIX -x
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
	if *printDirectories {
		run.conditions = append(run.conditions, isDirectory)
	}
	if *printFiles {
		run.conditions = append(run.conditions, isNotDirectory)
	}
	if *hideHidden {
		run.conditions = append(run.conditions, isNotHidden)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	if *printAbsolute {
		run.print = func(path string) {
			printAbsoluteFile(path)
		}
	} else {
		run.print = func(path string) {
			for _, condition := range run.conditions {
				if !condition(path) {
					return
				}
			}
			if !*noRelativePrefix && !strings.HasPrefix(path, "./") && !strings.HasPrefix(path, "../") {
				path = "./" + path
			}
			fmt.Println(path)
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
	dir := getDirectory(path, depth)
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

// getDirectory reads the directory and returns its contents
func getDirectory(path string, level int64) *Directory {
	files, err := os.ReadDir(path)
	if err != nil || files == nil {
		printError(err)
		return nil
	}
	return &Directory{Name: path, Files: files, Level: level}
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
