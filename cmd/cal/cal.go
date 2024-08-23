// Copyright (c) 2024-2024 xplshn						[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	reset = "\033[0m"
	red   = "\033[41m"
)

var (
	months = [12]string{
		"January", "February", "March", "April",
		"May", "June", "July", "August",
		"September", "October", "November", "December",
	}
	daysSundayFirst = [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	daysMondayFirst = [7]string{"Mo", "Tu", "We", "Th", "Fr", "Sa", "Su"}
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Name:        "cal",
		Synopsis:    "<|-h|-j|-m|-w|-y|> <month> <year>",
		Description: "Displays a calendar",
		CustomFields: map[string]interface{}{
			"Notes": "This version of cal contains does not account for the Gregorian Reformation that happened in 1752 after the 2nd of September.\nThis implementation is pure technical debt and is NOT conformant to https://man.openbsd.org/cal.1\nThat's going to change soon I hope.",
		},
	}

	julian := flag.Bool("j", false, "Use Julian dates")
	monday := flag.Bool("m", false, "Week starts on Monday")
	yearly := flag.Bool("y", false, "Display the entire year")
	weekNum := flag.Bool("w", false, "Display week numbers")
	noHighlight := flag.Bool("h", false, "Don't highlight current date")

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()
	args := flag.Args()

	if *julian && *weekNum {
		fmt.Println("Error: -j and -w options are mutually exclusive.")
		os.Exit(1)
	}

	if len(args) > 3 {
		fmt.Println("Usage: cal [-jmywh] [[[DAY] MONTH] YEAR]")
		return
	}

	var year, month, day int
	switch len(args) {
	case 0:
		t := time.Now()
		year = t.Year()
		month = int(t.Month())
		day = t.Day()
	case 1:
		year = parseInt(args[0])
		if year < 1 || year > 9999 {
			fmt.Println("Error: Invalid year")
			os.Exit(1)
		}
		printYear(year, *julian, *monday, *weekNum, *noHighlight, day, month)
		return
	case 2:
		month = parseMonth(args[0])
		year = parseInt(args[1])
		if year < 1 || year > 9999 {
			fmt.Println("Error: Invalid year")
			os.Exit(1)
		}
		if month < 1 || month > 12 {
			fmt.Println("Error: Invalid month")
			os.Exit(1)
		}
	case 3:
		day = parseInt(args[0])
		month = parseMonth(args[1])
		year = parseInt(args[2])
		if year < 1 || year > 9999 {
			fmt.Println("Error: Invalid year")
			os.Exit(1)
		}
		if month < 1 || month > 12 {
			fmt.Println("Error: Invalid month")
			os.Exit(1)
		}
		if day < 1 || day > 31 {
			fmt.Println("Error: Invalid day")
			os.Exit(1)
		}
	}

	if *yearly {
		printYear(year, *julian, *monday, *weekNum, *noHighlight, day, month)
	} else if month == 0 {
		printMonth(int(time.Now().Month()), year, *julian, *monday, *weekNum, *noHighlight, day)
	} else {
		printMonth(month, year, *julian, *monday, *weekNum, *noHighlight, day)
	}
}

func printMonth(month, year int, julian, monday, weekNum, noHighlight bool, highlightDay int) {
	content := calculateMonthContent(month, year, julian, monday, weekNum, noHighlight, highlightDay)
	header := fmt.Sprintf("%s %d", months[month-1], year)
	fmt.Println(ccmd.CFormatCenter(header, ccmd.RelativeTo(content)))
	fmt.Println(content)
}

func calculateMonthContent(month, year int, julian, monday, weekNum, noHighlight bool, highlightDay int) string {
	var content strings.Builder
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	lastDayOfMonth := firstDayOfMonth.AddDate(0, 1, -1)
	weekDay := int(firstDayOfMonth.Weekday())

	if monday {
		weekDay = (weekDay + 6) % 7
	}

	if julian {
		if monday {
			content.WriteString(" Mo  Tu  We  Th  Fr  Sa  Su\n")
		} else {
			content.WriteString(" Su  Mo  Tu  We  Th  Fr  Sa\n")
		}
	} else {
		if monday {
			content.WriteString(strings.Join(daysMondayFirst[:], " ") + "\n")
		} else {
			content.WriteString(strings.Join(daysSundayFirst[:], " ") + "\n")
		}
	}

	var lines []string
	var currentLine string

	for day := 1; day <= lastDayOfMonth.Day(); day++ {
		if julian {
			julianDay := firstDayOfMonth.YearDay() + day - 1
			if !noHighlight && day == highlightDay && month == int(time.Now().Month()) {
				currentLine += fmt.Sprintf("%s%3d%s ", red, julianDay, reset)
			} else {
				currentLine += fmt.Sprintf("%3d ", julianDay)
			}
		} else {
			if !noHighlight && day == highlightDay && month == int(time.Now().Month()) {
				currentLine += fmt.Sprintf("%s%2d%s ", red, day, reset)
			} else {
				currentLine += fmt.Sprintf("%2d ", day)
			}
		}

		if (weekDay+day)%7 == 0 {
			if weekNum {
				weekNumber := getWeekNumber(day, month, year, monday)
				currentLine += fmt.Sprintf(" [%2d]", weekNumber)
			}
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}

	if currentLine != "" {
		lines = append(lines, currentLine)
	}

	restOfLines := strings.Join(lines[1:], "\n")
	content.WriteString(ccmd.CFormatRight(lines[0], ccmd.RelativeTo(restOfLines)) + "\n")
	content.WriteString(restOfLines + "\n")

	return content.String()
}

func getWeekNumber(day, month, year int, monday bool) int {
	firstDayOfYear := time.Date(year, time.January, 1, 0, 0, 0, 0, time.Local)
	firstDayOfMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.Local)
	currentDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

	if monday {
		_, week := currentDate.ISOWeek()
		return week
	} else {
		daysSinceStartOfYear := currentDate.Sub(firstDayOfYear).Hours() / 24
		weekNumber := int(daysSinceStartOfYear/7) + 1
		if firstDayOfMonth.Weekday() == time.Sunday {
			weekNumber++
		}
		return weekNumber
	}
}

func printYear(year int, julian, monday, weekNum, noHighlight bool, highlightDay, highlightMonth int) {
	fmt.Printf("                               %d\n\n", year)
	if julian {
		printTwoMonths(1, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
		printTwoMonths(3, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
		printTwoMonths(5, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
		printTwoMonths(7, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
		printTwoMonths(9, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
		printTwoMonths(11, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, true)
	} else {
		for m := 1; m <= 12; m += 3 {
			printThreeMonths(m, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth)
		}
	}
}

func printTwoMonths(startMonth, year int, julian, monday, weekNum, noHighlight bool, highlightDay, highlightMonth int, yearly bool) {
	monthStrings := make([][]string, 2)
	for i := 0; i < 2; i++ {
		monthStrings[i] = getMonthStrings(startMonth+i, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, yearly)
	}

	maxLines := 0
	for i := 0; i < 2; i++ {
		if len(monthStrings[i]) > maxLines {
			maxLines = len(monthStrings[i])
		}
	}

	for line := 0; line < maxLines; line++ {
		for i := 0; i < 2; i++ {
			if line < len(monthStrings[i]) {
				fmt.Printf("%-35s", monthStrings[i][line])
			} else {
				fmt.Printf("%-35s", "")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func printThreeMonths(startMonth, year int, julian, monday, weekNum, noHighlight bool, highlightDay, highlightMonth int) {
	monthStrings := make([][]string, 3)
	for i := 0; i < 3; i++ {
		monthStrings[i] = getMonthStrings(startMonth+i, year, julian, monday, weekNum, noHighlight, highlightDay, highlightMonth, false)
	}

	maxLines := 0
	for i := 0; i < 3; i++ {
		if len(monthStrings[i]) > maxLines {
			maxLines = len(monthStrings[i])
		}
	}

	for line := 0; line < maxLines; line++ {
		for i := 0; i < 3; i++ {
			if line < len(monthStrings[i]) {
				fmt.Printf("%-22s", monthStrings[i][line])
			} else {
				fmt.Printf("%-22s", "")
			}
		}
		fmt.Println()
	}
}

func getMonthStrings(month, year int, julian, monday, weekNum, noHighlight bool, highlightDay, highlightMonth int, yearly bool) []string {
	var lines []string
	content := calculateMonthContent(month, year, julian, monday, weekNum, noHighlight, highlightDay)
	header := fmt.Sprintf("%s", months[month-1])
	maxLength := ccmd.RelativeTo(content)
	centeredHeader := ccmd.CFormatCenter(header, maxLength)
	lines = append(lines, centeredHeader)
	lines = append(lines, strings.Split(content, "\n")...)
	return lines
}

func parseInt(str string) int {
	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return val
}

func parseMonth(str string) int {
	for i, m := range months {
		if strings.EqualFold(m, str) {
			return i + 1
		}
	}
	val, err := strconv.Atoi(str)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return val
}
