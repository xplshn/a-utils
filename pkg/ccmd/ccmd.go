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
	Options      []string               // Populated automatically
	ExcludeFlags map[string]bool        // Tracks flags to exclude from help
	Since        int                    // Start year of the project
	CustomFields map[string]interface{} // Support for additional custom fields
}

// PopulateOptions fills the Options slice based on registered flags.
func (ci *CmdInfo) PopulateOptions() {
	ci.Options = nil
	flag.VisitAll(func(f *flag.Flag) {
		if !ci.ExcludeFlags[f.Name] {
			prefix := "--"
			if len(f.Name) == 1 {
				prefix = "-"
			}
			ci.Options = append(ci.Options, fmt.Sprintf("%s%s: %s", prefix, f.Name, f.Usage))
		}
	})
}

// GenerateHelpPage creates a help page based on CmdInfo fields.
func (ci *CmdInfo) GenerateHelpPage() (string, error) {
	if ci.Name == "" || ci.Description == "" || (ci.Synopsis == "" && ci.Usage == "") {
		return "", fmt.Errorf("Name, Description, and either Synopsis or Usage must be set")
	}
	ci.PopulateOptions()

	sb := &strings.Builder{}

	// Copyright and Authors
	year := time.Now().Year()
	if ci.Since > 0 {
		sb.WriteString(fmt.Sprintf("\n Copyright (c) %d-%d: ", ci.Since, year))
	} else {
		sb.WriteString(fmt.Sprintf("\n Copyright (c) %d: ", year))
	}
	sb.WriteString(strings.Join(ci.Authors, ", ") + " and contributors\n")
	if ci.Repository != "" {
		sb.WriteString(fmt.Sprintf(" For more details refer to %s\n", ci.Repository))
	}

	// Synopsis or Usage
	if ci.Synopsis != "" {
		sb.WriteString("\n  Synopsis\n")
		sb.WriteString(fmt.Sprintf("    %s %s\n", ci.Name, ci.Synopsis))
	} else if ci.Usage != "" {
		sb.WriteString("\n  Usage\n")
		sb.WriteString(fmt.Sprintf("    %s %s\n", ci.Name, ci.Usage))
	}

	// Description
	sb.WriteString("  Description:\n")
	sb.WriteString(fmt.Sprintf("    %s\n", ci.Description))

	// Options
	if len(ci.Options) > 0 {
		sb.WriteString("  Options:\n")
		for _, opt := range ci.Options {
			sb.WriteString(fmt.Sprintf("    %s\n", opt))
		}
	}

	// Custom Fields
	for i := 1; i <= len(ci.CustomFields); i++ {
		key := fmt.Sprintf("%d_", i)
		for field, value := range ci.CustomFields {
			if strings.HasPrefix(field, key) {
				sb.WriteString(fmt.Sprintf("  %s:\n", strings.TrimPrefix(field, key)))

				// Ensure that each line is indented the same as the first line
				lines := strings.Split(fmt.Sprintf("%s", value), "\n")
				for _, line := range lines {
					sb.WriteString(fmt.Sprintf("    %s\n", line))
				}
			}
		}
	}

	// Notes (always the last field)
	if notes, ok := ci.CustomFields["Notes"]; ok {
		sb.WriteString("  Notes:\n")
		lines := strings.Split(fmt.Sprintf("%s", notes), "\n")
		for _, line := range lines {
			sb.WriteString(fmt.Sprintf("    %s\n", line))
		}
	}

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
