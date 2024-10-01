// walk traverses directory trees, printing visited files // heavily modified version of ![Torgo's walk.go](https://github.com/as/torgo/blob/master/walk/walk.go)
package main

import (
	"bufio"
	"flag"
	"fmt"
	"strings"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const Prefix = "walk: "

// Constants and global variables
var (
	NGo   = 1024 * 16
	gosem = make(chan struct{}, NGo)
)

var run struct {
	traverse func(string, func(string))
	conditions []func(string) bool
	print func(string)
}

type Directory struct {
	Name  string
	Files []os.FileInfo
	Level int64
}

var (
	visitedMap = make(map[string]bool)
	rwlock      sync.RWMutex
	visitedFunc = markVisited
)

// isDirectory checks if the given path is a directory
func isDirectory(f string) bool {
	fi, err := os.Stat(f)
	if err != nil {
		printError(err)
		return false
	}
	return fi.IsDir()
}

// isNotDirectory checks if the given path is not a directory
func isNotDirectory(f string) bool {
	return !isDirectory(f)
}

// isNotHidden checks if the given path is not a hidden file
func isNotHidden(e string) bool {
	return !strings.HasPrefix(e, ".")
}

// markVisited marks a file as visited
func markVisited(f string) (yes bool) {
    rwlock.RLock()
    _, yes = visitedMap[f]
    rwlock.RUnlock()

    if !yes {
        rwlock.Lock()
        visitedMap[f] = true
        rwlock.Unlock()
    }
    return
}

// printFile prints the file path if it meets all conditions
func printFile(f string) {
	for _, condition := range run.conditions {
		if !condition(f) {
			return
		}
	}
	fmt.Println(f)
}

// printAbsoluteFile prints the absolute path of the file if it meets all conditions
func printAbsoluteFile(f string) bool {
	var err error
	if !filepath.IsAbs(f) {
		if f, err = filepath.Abs(f); err != nil {
			printError(err)
			return false
		}
	}
	printFile(f)
	return true
}

var args struct {
	h, q             bool
	a, d, f, x       bool
	t                int64
}

func main() {
	flag.BoolVar(&args.a, "a", false, "Print absolute paths")
	flag.BoolVar(&args.d, "d", false, "Print directories only")
	flag.BoolVar(&args.f, "f", false, "Print files only")
	flag.BoolVar(&args.x, "x", false, "Hide hidden files")
	flag.Int64Var(&args.t, "t", 1024*1024*1024, "Traversal limit")

	cmdInfo := &ccmd.CmdInfo{
		Name:        "walk",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "traverse a list of targets (directories or files)",
		Synopsis:    "<|-a|-t [INT]|> <|-d|-f|-x|> [target ...]",
		CustomFields: map[string]interface{}{
			"1_Behavior": `Walk walks the named file list and prints each name
to standard output. A directory in the file list is
a file list. The file "-" names standard input as a
file list of line-separated file names.`,
			"2_Example": `Walk the first four levels down the directory
tree, look for "mobius", and walk all files in
the mobius directories.
	\$ walk -d -t 4 | grep -i mobius | walk -f -`,
			"3_Notes":    "Walk will not follow symlinks",
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
	flag.Parse()

	if args.d && args.f {
		printError("bad args: -d and -f")
		os.Exit(1)
	}

	run.traverse = depthFirstTraversal
	if args.d {
		run.conditions = append(run.conditions, isDirectory)
	}
	if args.f {
		run.conditions = append(run.conditions, isNotDirectory)
	}
	if args.x {
		run.conditions = append(run.conditions, isNotHidden)
	}

	paths := flag.Args()
	if len(paths) == 0 {
		paths = []string{"."} // Current working dir
		run.print = func(f string) {
			for _, condition := range run.conditions {
				if !condition(f) {
					return
				}
			}
			if len(f) > 2 {
				fmt.Println(f[2:])
			}
		}
	}
	if args.a {
		run.print = func(f string) {
			var err error
			for _, condition := range run.conditions {
				if !condition(f) {
					return
				}
			}
			if f, err = filepath.Abs(f); err != nil {
				printError(err)
				return
			}
			fmt.Println(f)
		}
	} else {
		run.print = printFile
	}

	var wg sync.WaitGroup
	for _, v := range paths {
		wg.Add(1)
		go func(v string) {
			defer wg.Done()
			if v != "-" {
				run.traverse(v, run.print)
			} else {
				in := bufio.NewScanner(os.Stdin)
				for in.Scan() {
					run.traverse(in.Text(), run.print)
				}
			}
		}(v)
	}
	wg.Wait()
}

// depthFirstTraversal performs a depth-first traversal of the file f
func depthFirstTraversal(name string, fn func(string)) {
	listch := make(chan string, NGo)
	var wg sync.WaitGroup
	wg.Add(1)
	go depthFirstTraversalHelper(name, &wg, 0, listch)
	go func() {
		wg.Wait()
		close(listch)
	}()
	for d := range listch {
		run.print(d)
	}
}

func depthFirstTraversalHelper(nm string, wg *sync.WaitGroup, deep int64, listch chan<- string) {
	defer wg.Done()
	d := getDirectory(nm, deep)
	if d == nil {
		return
	}
	for _, f := range d.Files {
		kid := filepath.Join(d.Name, f.Name())
		if visitedFunc(kid) {
			continue
		}
		if f.IsDir() && deep < args.t {
			wg.Add(1)
			go depthFirstTraversalHelper(kid, wg, deep+1, listch)
		}
		listch <- kid
	}
}

// getDirectory reads the directory and returns its contents
func getDirectory(n string, level int64) (dir *Directory) {
	f, err := ioutil.ReadDir(n)
	if err != nil || f == nil {
		printError(err)
		return nil
	}
	return &Directory{n, f, level}
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
