// walk traverses directory trees, printing visited files // heavily modified version of ![Torgo's walk.go](https://github.com/as/torgo/blob/master/walk/walk.go)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/charlievieth/fastwalk"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

const Prefix = "walk: "

var (
	visitedMap  = make(map[string]bool)
	rwlock      sync.RWMutex
	visitedFunc = markVisited
)

type FileInfo struct {
	Path      string
	IsDir     bool
	IsSymlink bool
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

func isHidden(path string) bool {
	components := strings.Split(path, "/")
	filename := components[len(components)-1]
	if strings.HasPrefix(filename, ".") {
		return true
	}
	for _, component := range components {
		if strings.HasPrefix(component, ".") {
			return true
		}
	}
	return false
}

func isNotHidden(path string) bool {
	return !isHidden(path)
}

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

func printFile(path string, conditions []func(string) bool) {
	for _, condition := range conditions {
		if !condition(path) {
			return
		}
	}
	fmt.Println(path)
}

func printAbsoluteFile(path string, conditions []func(string) bool) bool {
	var err error
	if !filepath.IsAbs(path) {
		if path, err = filepath.Abs(path); err != nil {
			printError(err)
			return false
		}
	}
	printFile(path, conditions)
	return true
}

func main() {
	traversalLimit := flag.Int64("t", 1024*1024*1024, "Traversal depth limit")
	printDirectories := flag.Bool("d", false, "Print directories only")
	printFiles := flag.Bool("f", false, "Print files only")
	printAbsolute := flag.Bool("A", false, "Print absolute paths")
	showAll := flag.Bool("a", false, `Show paths that contain a directory or file prepended with '.'`)

	cmdInfo := &ccmd.CmdInfo{
		Name:        "walk",
		Authors:     []string{"as", "xplshn"}, // Should his name ("as") be here? This is nothing like the original, not anymore.
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "traverse a list of targets (directories or files)",
		Synopsis:    "<|-t [INT]|-d|-f|-A|-a|> [target ...]",
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

	var conditions []func(string) bool
	if !*showAll {
		conditions = append(conditions, isNotHidden)
	}
	if *printDirectories {
		conditions = append(conditions, isdirectory)
	}
	if *printFiles {
		conditions = append(conditions, isNotdirectory)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."}
	}

	var wg sync.WaitGroup
	var visitedCount int64
	var visitedLock sync.Mutex

	for _, target := range paths {
		wg.Add(1)
		go func(target string) {
			defer wg.Done()
			if target != "-" {
				walkFn := func(path string, d fs.DirEntry, err error) error {
					if err != nil {
						fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
						return nil
					}
					if visitedFunc(path) {
						return nil
					}
					visitedLock.Lock()
					if visitedCount >= *traversalLimit {
						visitedLock.Unlock()
						return fs.SkipAll
					}
					visitedCount++
					visitedLock.Unlock()
					if *printAbsolute {
						printAbsoluteFile(path, conditions)
					} else {
						printFile(path, conditions)
					}
					return nil
				}
				conf := fastwalk.Config{
					Follow: false,
				}
				if err := fastwalk.Walk(&conf, target, walkFn); err != nil {
					fmt.Fprintf(os.Stderr, "%s: %v\n", target, err)
					os.Exit(1)
				}
			} else {
				in := bufio.NewScanner(os.Stdin)
				for in.Scan() {
					walkFn := func(path string, d fs.DirEntry, err error) error {
						if err != nil {
							fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
							return nil
						}
						if visitedFunc(path) {
							return nil
						}
						visitedLock.Lock()
						if visitedCount >= *traversalLimit {
							visitedLock.Unlock()
							return fs.SkipAll
						}
						visitedCount++
						visitedLock.Unlock()
						if *printAbsolute {
							printAbsoluteFile(path, conditions)
						} else {
							printFile(path, conditions)
						}
						return nil
					}
					conf := fastwalk.Config{
						Follow: false,
					}
					if err := fastwalk.Walk(&conf, in.Text(), walkFn); err != nil {
						fmt.Fprintf(os.Stderr, "%s: %v\n", in.Text(), err)
						os.Exit(1)
					}
				}
			}
		}(target)
	}
	wg.Wait()
}

func println(v ...interface{}) {
	fmt.Print(Prefix)
	fmt.Println(v...)
}

func printError(v ...interface{}) {
	fmt.Fprint(os.Stderr, Prefix)
	fmt.Fprintln(os.Stderr, v...)
}
