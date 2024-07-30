// Consistent Command Line
package ccmd

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type CmdInfo struct {
	Repository   string
	Authors      []string
	Name         string
	Synopsis     string // Either Synopsis or Usage must be set
	Usage        string // ------------------------------------
	Description  string
	Behavior     string          // Optional
	Options      []string        // Optional
	Notes        string          // Optional
	ExcludeFlags map[string]bool // Tracks flags to exclude from help
	Since        int             // When the project started. Used to print copyright as: (c) 2017-2024: xplshn and contributors
}

func (ci *CmdInfo) DefineFlag(name string, value interface{}, usage string, exclude bool) {
	// Define the flag using the standard flag package
	switch v := value.(type) {
	case bool:
		flag.Bool(name, v, usage)
	case *bool:
		flag.BoolVar(v, name, *v, usage)
	case int:
		flag.Int(name, v, usage)
	case *int:
		flag.IntVar(v, name, *v, usage)
	case string:
		flag.String(name, v, usage)
	case *string:
		flag.StringVar(v, name, *v, usage)
	case float64:
		flag.Float64(name, v, usage)
	case *float64:
		flag.Float64Var(v, name, *v, usage)
	case time.Duration:
		flag.Duration(name, v, usage)
	case *time.Duration:
		flag.DurationVar(v, name, *v, usage)
	default:
		panic("unsupported flag type")
	}

	// Track the flag and its exclusion status
	ci.ExcludeFlags[name] = exclude
}

func (ci *CmdInfo) PopulateOptions() {
	// Clear existing options
	ci.Options = nil

	flag.VisitAll(func(f *flag.Flag) {
		// Determine if the flag is long or shorthand
		var flagFormat string
		if len(f.Name) == 1 {
			flagFormat = fmt.Sprintf("-%s", f.Name)
		} else {
			flagFormat = fmt.Sprintf("--%s", f.Name)
		}
		if !ci.ExcludeFlags[f.Name] {
			ci.Options = append(ci.Options, fmt.Sprintf("%s: %s", flagFormat, f.Usage))
		}
	})
}

func (ci *CmdInfo) GenerateHelpPage() (string, error) {
	if ci.Name == "" || ci.Description == "" || (ci.Synopsis == "" && ci.Usage == "") {
		return "", fmt.Errorf("mandatory fields missing: Name, Description, and either Synopsis or Usage must be set")
	}

	// Populate the options automatically based on the flags defined
	ci.PopulateOptions()

	var sb strings.Builder

	// Copyright
	if ci.Since > 0 {
		sb.WriteString(fmt.Sprintf("\n Copyright (c) %d-%d: ", ci.Since, time.Now().Year()))
	} else {
		sb.WriteString(fmt.Sprintf("\n Copyright (c) %d: ", time.Now().Year()))
	}
	for _, author := range ci.Authors {
		sb.WriteString(author + ", ")
	}
	sb.WriteString("and contributors\n")
	if ci.Repository == "" {
		sb.WriteString(" For more details refer to https://github.com/xplshn/a-utils\n")
	} else {
		sb.WriteString(fmt.Sprintf(" For more details refer to %s\n", ci.Repository))
	}

	// Synopsis or Usage
	if ci.Synopsis != "" {
		sb.WriteString("\n  Synopsis\n")
		sb.WriteString(fmt.Sprintf("    %s %s\n", ci.Name, ci.Synopsis))
	} else if ci.Usage != "" {
		sb.WriteString("\n  Synopsis\n")
		sb.WriteString(fmt.Sprintf("    %s %s\n", ci.Name, ci.Usage))
	}

	// Description
	sb.WriteString("  Description:\n")
	sb.WriteString(fmt.Sprintf("    %s\n", ci.Description))

	// Behavior
	if ci.Behavior != "" {
		sb.WriteString("  Behavior:\n")
		sb.WriteString(fmt.Sprintf("    %s\n", ci.Behavior))
	}

	// Options
	if len(ci.Options) > 0 {
		sb.WriteString("  Options:\n")
		for _, opt := range ci.Options {
			sb.WriteString(fmt.Sprintf("    %s\n", opt))
		}
	}

	// Notes
	if ci.Notes != "" {
		noteLines := strings.Split(ci.Notes, "\n")
		if len(noteLines) > 1 {
			sb.WriteString("  Notes:\n")
		} else {
			sb.WriteString("  Note:\n")
		}
		for _, line := range noteLines {
			sb.WriteString(fmt.Sprintf("    %s\n", line))
		}
	}

	// Ensure there is a newline at the end of the help page
	sb.WriteString("\n")

	return sb.String(), nil
}

// GetTerminalWidth attempts to determine the width of the terminal.
// It first tries using "stty size", then "tput cols", and finally falls back to 80 columns.
func GetTerminalWidth() int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(out)), " ")
		if len(parts) == 2 {
			width, _ := strconv.Atoi(parts[1])
			return width
		}
	}

	cmd = exec.Command("tput", "cols")
	cmd.Stdin = os.Stdin
	out, err = cmd.Output()
	if err == nil {
		width, _ := strconv.Atoi(strings.TrimSpace(string(out)))
		return width
	}

	return 80
}

// CFormatCenter formats the text to be centered based on the width passed by the user.
func CFormatCenter(text string, width int) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder

	for _, line := range lines {
		lineLength := len(line)
		if lineLength < width {
			padding := (width - lineLength) / 2
			// Print the line padded to the proper width
			sb.WriteString(fmt.Sprintf("%*s", width-padding, line))
		} else {
			sb.WriteString(line)
		}
	}

	return sb.String()
}

// FormatCenter formats the text to be centered based on the terminal width.
func FormatCenter(text string) string {
	return CFormatCenter(text, GetTerminalWidth())
}

// FormatRight formats the text to be right-aligned based on the terminal width.
func FormatRight(text string) string {
	return CFormatRight(text, GetTerminalWidth())
}

// CFormatRight formats the text to be right-aligned based on the width passed by the user.
func CFormatRight(text string, width int) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder

	for _, line := range lines {
		lineLength := len(line)
		if lineLength < width {
			// Right-align the line with padding on the left
			sb.WriteString(fmt.Sprintf("%*s", width, line))
		} else {
			sb.WriteString(line)
		}
	}

	return sb.String()
}

// FormatLeft formats the text to be left-aligned based on the terminal width.
func FormatLeft(text string) string {
	return CFormatLeft(text, GetTerminalWidth())
}

// CFormatLeft formats the text to be left-aligned based on the width passed by the user.
func CFormatLeft(text string, width int) string {
	lines := strings.Split(text, "\n")
	var sb strings.Builder

	for _, line := range lines {
		lineLength := len(line)
		if lineLength < width {
			sb.WriteString(fmt.Sprintf("%-*s", width, line))
		} else {
			sb.WriteString(line)
		}
	}

	return sb.String()
}

// RelativeTo finds the length of the longest line in the given string, accounting for non-printing characters and ANSI escape sequences.
func RelativeTo(input string) int {
	// Regular expression to match ANSI escape sequences
	ansiEscape := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

	lines := strings.Split(input, "\n")
	maxLength := 0
	for _, line := range lines {
		// Remove ANSI escape sequences
		cleanLine := ansiEscape.ReplaceAllString(line, "")
		lineLength := utf8.RuneCountInString(cleanLine)
		if lineLength > maxLength {
			maxLength = lineLength
		}
	}
	return maxLength
}
