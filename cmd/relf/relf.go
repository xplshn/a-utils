package main

import (
	"bufio"
	"debug/elf"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

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

// Refinement holds functions to check for specific conditions and the corresponding section name.
type Refinement struct {
	SectionCheck func(*elf.File) bool
	StringCheck  func(string) bool
	Name         string
}

// containsString checks if the file contains a specific string using the strings utility logic.
func containsString(filename, target string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	in := bufio.NewReader(file)
	str := make([]rune, 0, 256)
	filePos := int64(0)
	for {
		var (
			r   rune
			wid int
			err error
		)
		for {
			r, wid, err = in.ReadRune()
			if err != nil {
				if err != io.EOF {
					fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
				}
				return false
			}
			filePos += int64(wid)
			if !strconv.IsPrint(r) || r >= 0xFF {
				if strings.Contains(string(str), target) {
					return true
				}
				str = str[0:0]
				continue
			}
			if len(str) >= 256 {
				if strings.Contains(string(str), target) {
					return true
				}
				str = str[0:0]
			}
			str = append(str, r)
		}
	}
}

// formatSize formats the size based on the human-readable flag.
func formatSize(size uint64, humanReadable bool) string {
	if humanReadable {
		return humanReadableSize(size)
	}
	return fmt.Sprintf("%d", size)
}

// printSectionSizes prints the sizes of each section in a human-readable format.
func printSectionSizes(file *elf.File, filePath string, fileSize uint64, refinements []Refinement, humanReadable bool) {
	var totalSize uint64
	var sections []*elf.Section
	for _, section := range file.Sections {
		if section.Size > 0 {
			sections = append(sections, section)
			totalSize += section.Size
		}
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Size > sections[j].Size
	})

	for _, section := range sections {
		fmt.Printf("%-24s %s\n", section.Name, formatSize(section.Size, humanReadable))
	}

	if totalSize < fileSize {
		unknownSize := fileSize - totalSize
		sectionName := "unknown"
		for _, refinement := range refinements {
			if refinement.SectionCheck != nil && refinement.SectionCheck(file) {
				sectionName = refinement.Name
				break
			}
			if refinement.StringCheck != nil && refinement.StringCheck(filePath) {
				sectionName = refinement.Name
				break
			}
		}
		fmt.Printf("!%-24s %s\n", sectionName, formatSize(unknownSize, humanReadable))
	}
}

// printTree prints the tree structure of the sections.
func printTree(file *elf.File, filePath string, fileSize uint64, refinements []Refinement, humanReadable bool) {
	var totalSize uint64
	var sections []*elf.Section
	for _, section := range file.Sections {
		if section.Size > 0 {
			sections = append(sections, section)
			totalSize += section.Size
		}
	}

	sort.Slice(sections, func(i, j int) bool {
		return sections[i].Size > sections[j].Size
	})

	fmt.Println(filePath)
	for i, section := range sections {
		prefix := "├──"
		if i == len(sections)-1 {
			prefix = "└──"
		}
		fmt.Printf("    %s %s (%s)\n", prefix, section.Name, formatSize(section.Size, humanReadable))
	}

	if totalSize < fileSize {
		unknownSize := fileSize - totalSize
		sectionName := "!unknown"
		for _, refinement := range refinements {
			if refinement.SectionCheck != nil && refinement.SectionCheck(file) {
				sectionName = refinement.Name
				break
			}
			if refinement.StringCheck != nil && refinement.StringCheck(filePath) {
				sectionName = refinement.Name
				break
			}
		}
		fmt.Printf("    %s %s (%s)\n", "└──", sectionName, formatSize(unknownSize, humanReadable))
	}
}

// isElfFile checks if the file is an ELF file by reading its magic number.
func isElfFile(filename string) bool {
	file, err := os.Open(filename)
	if err != nil {
		return false
	}
	defer file.Close()

	var magic [4]byte
	_, err = file.Read(magic[:])
	if err != nil {
		return false
	}

	return string(magic[:]) == "\x7FELF"
}

func main() {
	treeFlag := flag.Bool("t", false, "Show the section sizes in a tree structure")
	humanReadableFlag := flag.Bool("s", false, "Show sizes in human-readable format")

	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Name:        "relf",
		Synopsis:    "<|-t|-s|> <elf-file>",
		Description: "Prints section sizes of ELF files",
		Repository:  "https://github.com/xplshn/a-utils",
		CustomFields: map[string]interface{}{
			"1_Examples": `Print section sizes in human-readable format:
  \$ relf -s file.elf`,
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

	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Print(helpPage)
		os.Exit(1)
	}

	filePath := flag.Arg(0)
	if !isElfFile(filePath) {
		fmt.Fprintf(os.Stderr, "Error: %s is not an ELF file\n", filePath)
		os.Exit(1)
	}

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting file info: %v\n", err)
		os.Exit(1)
	}
	fileSize := uint64(fileInfo.Size())

	refinements := []Refinement{
		{SectionCheck: func(file *elf.File) bool { return hasSection(file, ".appimage") }, StringCheck: func(filename string) bool {
			return containsString(filename, "--appimage-offset") && containsString(filename, "fsname=squashfuse")
		}, Name: "appimage.archive.sqfs"},
		{SectionCheck: func(file *elf.File) bool { return hasSection(file, ".sqfs") }, StringCheck: func(filename string) bool { return containsString(filename, "fsname=squashfuse") }, Name: "sqfs"},
		//{SectionCheck: func(file *elf.File) bool { return hasSection(file, ".dwfs") }, StringCheck: func(filename string) bool { return containsString(filename, "DWARFS") }, Name: "!dwfs"}, // unreliable
	}

	file, err := elf.Open(filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	if *treeFlag {
		printTree(file, filePath, fileSize, refinements, *humanReadableFlag)
	} else {
		printSectionSizes(file, filePath, fileSize, refinements, *humanReadableFlag)
	}
}

// hasSection checks if the ELF file contains a specific section.
func hasSection(file *elf.File, sectionName string) bool {
	for _, section := range file.Sections {
		if section.Name == sectionName {
			return true
		}
	}
	return false
}
