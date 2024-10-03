package main

import (
	"bufio"
	"flag"
	"strings"
	"fmt"
	"os"
	"time"
	"syscall"
	"path/filepath"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	Prefix = "fi: "
)

// Define colors and attributes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	AttrBold    = "\033[1m"
	AttrDim     = "\033[2m"
	AttrUnderline = "\033[4m"
	AttrBlink   = "\033[5m"
	AttrReverse = "\033[7m"
)

// FileData represents the information gathered for each file
type FileData struct {
	Path       string
	Mode       os.FileMode
	Inode      uint64
	Blocks     int64
	Size       uint64
	Nlink      uint64
	UID        uint32
	GID        uint32
	Mtime      time.Time
	Ctime      time.Time
	Atime      time.Time
	IsDir      bool
	IsSymlink  bool
}

// gatherFileData collects file information based on the provided path
func gatherFileData(path string) (FileData, error) {
	var fileInfo os.FileInfo
	var err error

	fileInfo, err = os.Stat(path)

	if err != nil {
		return FileData{}, err
	}

	sysStat, ok := fileInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return FileData{}, fmt.Errorf("failed to get raw syscall.Stat_t for file: %s", path)
	}

	return FileData{
		Path:       path,
		Mode:       fileInfo.Mode(),
		Inode:      sysStat.Ino,
		Blocks:     sysStat.Blocks,
		Size:       uint64(fileInfo.Size()),
		Nlink:      sysStat.Nlink,
		UID:        sysStat.Uid,
		GID:        sysStat.Gid,
		Mtime:      timespecToTime(sysStat.Mtim),
		Ctime:      timespecToTime(sysStat.Ctim),
		Atime:      timespecToTime(sysStat.Atim),
		IsDir:      fileInfo.IsDir(),
		IsSymlink:  fileInfo.Mode()&os.ModeSymlink != 0,
	}, nil
}

// Function to determine the color and attribute based on the file mode
func getColorAndAttr(mode os.FileMode) (string, string) {
	switch {
	case mode.IsDir():
		return ColorBlue, AttrBold
	case mode&os.ModeSymlink != 0:
		return ColorCyan, AttrBold
	case mode&os.ModeNamedPipe != 0:
		return ColorYellow, AttrBold
	case mode&os.ModeSocket != 0:
		return ColorGreen, AttrBold
	case mode&os.ModeCharDevice != 0:
		return ColorYellow, AttrBold
	case mode&os.ModeDevice != 0:
		return ColorYellow, AttrBold
	case mode&os.ModeSetuid != 0:
		return ColorRed, AttrBold
	case mode&os.ModeSetgid != 0:
		return ColorRed, AttrBold
	case mode&os.ModeSticky != 0:
		return ColorRed, AttrBold
	case mode.IsRegular() && mode&0111 != 0:
		return ColorGreen, AttrBold
	default:
		return ColorReset, AttrBold
	}
}

// Function to handle color options
func getColorizedOutput(fd FileData, colorOption string) string {
	if colorOption == "never" {
		return fmt.Sprintf("%s\n", filepath.Base(fd.Path))
	} else if !isTerminal() {
		return fmt.Sprintf("%s\n", filepath.Base(fd.Path))
	}

	color, attr := getColorAndAttr(fd.Mode)

	return fmt.Sprintf("%s%s%s%s\n", attr, color, filepath.Base(fd.Path), ColorReset)
}

// Check if the output is a terminal
func isTerminal() bool {
	info, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (info.Mode() & os.ModeCharDevice) != 0
}

func main() {
	showAll := flag.Bool("a", false, "Include names starting with .")
	showAlmostAll := flag.Bool("A", false, "Like -a, but exclude . and ..")
	appendSlash := flag.Bool("p", false, "Append / to directory names")
	appendIndicator := flag.Bool("F", false, "Append indicator (one of */=@|) to names")
	longFormat := flag.Bool("l", false, "Long format")
	showInode := flag.Bool("i", false, "Show inode numbers")
	showNumeric := flag.Bool("n", false, "Show numeric UIDs and GIDs instead of names")
	showBlocks := flag.Bool("s", false, "Show allocated blocks")
	showCtime := flag.Bool("lc", false, "Show ctime")
	showAtime := flag.Bool("lu", false, "Show atime")
	fullTime := flag.Bool("full-time", false, "List full date/time")
	humanReadable := flag.Bool("h", false, "Human readable sizes (1K 243M 2G)")
	colorOption := flag.String("color", "never", "Colorize the output: 'always', 'auto', 'never'")

	cmdInfo := &ccmd.CmdInfo{
		Name:        "fi",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Prints file information for files read from stdin or arguments",
		Synopsis:    "[-s -c -m -d -p -v -h --color=auto|always|never] [file1 [file2 ...]]",
		CustomFields: map[string]interface{}{
			"1_Examples": `Print file sizes and cumulative total:
  $ walk -f mink/ | fi -s -c`,
		},
	}
	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}
	// Set usage to print the CCMD-generated help page
	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	// Parse flags
	flag.Parse()

	var printFuncs []func(FileData) string

	// Build the list of printing functions based on the flags
	printFuncs = append(printFuncs, func(fd FileData) string {
		return getColorizedOutput(fd, *colorOption)
	})

	if *showInode {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%d\t", fd.Inode)
		})
	}
	if *showBlocks {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%d\t", fd.Blocks)
		})
	}
	if *longFormat {
		printFuncs = append(printFuncs, func(fd FileData) string {
			perm := fd.Mode.Perm().String()
			if fd.IsDir {
				perm = "d" + perm
			}
			if *showNumeric {
				return fmt.Sprintf("%s %d %d %d %d %s\t", perm, fd.Nlink, fd.UID, fd.GID, fd.Size, fd.Mtime.Format(time.RFC3339))
			} else {
				return fmt.Sprintf("%s %d %d %d %d %s\t", perm, fd.Nlink, fd.UID, fd.GID, fd.Size, fd.Mtime.Format(time.RFC3339))
			}
		})
	}
	if *showCtime {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%s\t", fd.Ctime.Format(time.RFC3339))
		})
	}
	if *showAtime {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%s\t", fd.Atime.Format(time.RFC3339))
		})
	}
	if *fullTime {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%s\t", fd.Mtime.Format(time.RFC3339))
		})
	}
	if *humanReadable {
		printFuncs = append(printFuncs, func(fd FileData) string {
			return fmt.Sprintf("%s\t", humanReadableSize(fd.Size))
		})
	}
	if *appendSlash {
		printFuncs = append(printFuncs, func(fd FileData) string {
			if fd.IsDir {
				return "/"
			}
			return ""
		})
	}
	if *appendIndicator {
		printFuncs = append(printFuncs, func(fd FileData) string {
			if fd.IsDir {
				return "/"
			} else if fd.IsSymlink {
				return "@"
			} else if fd.Mode&os.ModeNamedPipe != 0 {
				return "|"
			} else if fd.Mode&os.ModeSocket != 0 {
				return "="
			} else if fd.Mode&os.ModeCharDevice != 0 {
				return "*"
			} else if fd.Mode&os.ModeDevice != 0 {
				return "="
			}
			return ""
		})
	}

	totalSize := int64(0)

	// Check if any arguments are provided
	if len(flag.Args()) == 0 {
		// Check if stdin is being used
		stat, err := os.Stdin.Stat()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error checking stdin:", err)
			os.Exit(1)
		}
		if (stat.Mode() & os.ModeCharDevice) != 0 {
			// No arguments and stdin is not being used, show help page & exit
			fmt.Print(helpPage)
			fmt.Fprintln(os.Stderr, "No input files and stdin is not being used. Exiting.")
			os.Exit(1)
		}
	}

	// Process each file from arguments
	for _, fileName := range flag.Args() {
		fileData, err := gatherFileData(fileName)
		if err != nil {
			printError(err)
			continue
		}

		if !*showAll && !*showAlmostAll && strings.HasPrefix(filepath.Base(fileName), ".") {
			continue
		}
		if *showAlmostAll && (filepath.Base(fileName) == "." || filepath.Base(fileName) == "..") {
			continue
		}

		output := ""
		for i := len(printFuncs) - 1; i >= 0; i-- {
			output += printFuncs[i](fileData)
		}
		fmt.Print(output)
		totalSize += int64(fileData.Size)
	}

	// Check if stdin is being used
	stat, err := os.Stdin.Stat()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error checking stdin:", err)
		os.Exit(1)
	}
	if (stat.Mode() & os.ModeCharDevice) == 0 {
		// Process each file from stdin
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			fileName := scanner.Text()
			fileData, err := gatherFileData(fileName)
			if err != nil {
				printError(err)
				continue
			}

			if !*showAll && !*showAlmostAll && strings.HasPrefix(filepath.Base(fileName), ".") {
				continue
			}
			if *showAlmostAll && (filepath.Base(fileName) == "." || filepath.Base(fileName) == "..") {
				continue
			}

			output := ""
			for i := len(printFuncs) - 1; i >= 0; i-- {
				output += printFuncs[i](fileData)
			}
			fmt.Print(output)
			totalSize += int64(fileData.Size)
		}
		if err := scanner.Err(); err != nil {
			printError(err)
		}
	}

	// Print cumulative total if the flag is set
	if *showBlocks {
		fmt.Printf("%d total\n", totalSize)
	}
}

func printError(v ...interface{}) {
	fmt.Fprint(os.Stderr, Prefix)
	fmt.Fprintln(os.Stderr, v...)
}

// humanReadableSize converts a size in bytes to a human-readable format.
func humanReadableSize(size uint64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2f GB", float64(size)/GB)
	case size >= MB:
		return fmt.Sprintf("%.2f MB", float64(size)/MB)
	case size >= KB:
		return fmt.Sprintf("%.2f KB", float64(size)/KB)
	default:
		return fmt.Sprintf("%d B", size)
	}
}

// timespecToTime converts syscall.Timespec to time.Time
func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}
