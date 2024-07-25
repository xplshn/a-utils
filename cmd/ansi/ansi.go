// Copyright (c) 2024, xplshn, sweetbbak, u-root and contributors  [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strings"
)

var commands = map[string]string{
	// Special ANSI sequences
	"clear":               "\033[1;1H\033[2J",
	"clear_line":          "\r\x1b[K",
	"carriage_return":     "\r",
	"clear_cursor_to_eol": "\x1b[K",
	"hidecursor":          "\x1b[?25l",
	"showcursor":          "\x1b[?25h",
	"restore":             "\x1b[?47l",
	"save":                "\x1b[?47h",
	"rmaltscreen":         "\x1b[?1049l",
	"alt":                 "\x1b[?1049h",
	"reset_attributes":    "\x1b[0m",
	"reset":               "\033c\033(B\033[m\033[J\033[?25h",
	// Visible text attributes
	"__fg-red":            "\x1b[31m",
	"__fg-green":          "\x1b[32m",
	"__fg-yellow":         "\x1b[33m",
	"__fg-blue":           "\x1b[34m",
	"__fg-magenta":        "\x1b[35m",
	"__fg-cyan":           "\x1b[36m",
	"__fg-lightgrey":      "\x1b[37m",
	"__fg-darkgrey":       "\x1b[90m",
	"__fg-black":          "\x1b[30m",
	"__fg-lightred":       "\x1b[91m",
	"__fg-lightgreen":     "\x1b[92m",
	"__fg-lightyellow":    "\x1b[93m",
	"__fg-lightblue":      "\x1b[94m",
	"__fg-lightmagenta":   "\x1b[95m",
	"__fg-lightcyan":      "\x1b[96m",
	"__fg-white":          "\x1b[97m",
	"__bg-red":            "\x1b[41m",
	"__bg-green":          "\x1b[42m",
	"__bg-yellow":         "\x1b[43m",
	"__bg-blue":           "\x1b[44m",
	"__bg-magenta":        "\x1b[45m",
	"__bg-cyan":           "\x1b[46m",
	"__bg-lightgrey":      "\x1b[47m",
	"__bg-darkgrey":       "\x1b[40m",
	"__bg-black":          "\x1b[40m",
	"__bg-lightred":       "\x1b[101m",
	"__bg-lightgreen":     "\x1b[102m",
	"__bg-lightyellow":    "\x1b[103m",
	"__bg-lightblue":      "\x1b[104m",
	"__bg-lightmagenta":   "\x1b[105m",
	"__bg-lightcyan":      "\x1b[106m",
	"__bg-white":          "\x1b[107m",
	"__afg-bold":          "\x1b[1m",
	"__afg-dim":           "\x1b[2m",
	"__afg-italic":        "\x1b[3m",
	"__afg-underline":     "\x1b[4m",
	"__afg-blink":         "\x1b[5m",
	"__afg-reverse":       "\x1b[7m",
	"__afg-hidden":        "\x1b[8m",
	"__afg-strikethrough": "\x1b[9m",
}

func getCommand(key string) (string, bool) {
	if cmd, exists := commands[key]; exists {
		return cmd, true
	}
	if cmd, exists := commands["_"+key]; exists {
		return cmd, true
	}
	if cmd, exists := commands["__"+key]; exists {
		return cmd, true
	}
	return "", false
}

func ansi(w io.Writer, args []string, text string) error {
	for _, arg := range args {
		command, exists := getCommand(arg)
		if !exists {
			return fmt.Errorf("command ANSI '%v' doesn't exist", arg)
		}

		if strings.HasPrefix(arg, "__") && text != "" {
			fmt.Fprintf(w, "%s%s%s", command, text, commands["reset_attributes"])
		} else if strings.HasPrefix(arg, "_") {
			fmt.Fprintf(w, "%s%s", command, text)
		} else {
			fmt.Fprint(w, command)
		}
	}
	return nil
}

func main() {
	list := flag.Bool("l", false, "list available commands")
	text := flag.String("text", "", "Wrap text around attributes")
	flag.Usage = func() {
		p := `
 Copyright (c) 2024, xplshn, sweetbbak, u-root and contributors [3BSD]
 For more details refer to https://github.com/xplshn/a-utils

  Description
    Print ansi escape sequences.
  Synopsis:
    ansi <--text|--list|--output> [name]...
  Options:
    --text: wraps text around options that require a 'reset' at the end. Namely colors, among other attributes.
    --list: provides a list of the available ansi escape sequences.
`
		fmt.Println(p)
	}
	flag.Parse()

	if *list {
		keys := make([]string, 0, len(commands))
		for i := range commands {
			keys = append(keys, i)
		}
		sort.Strings(keys)
		for _, i := range keys {
			trimmed := strings.TrimPrefix(i, "__")
			trimmed = strings.TrimPrefix(trimmed, "_")
			fmt.Println(trimmed)
		}
		os.Exit(0)
	}

	var out *os.File = os.Stdout
	if err := ansi(out, flag.Args(), *text); err != nil {
		log.Fatalf("%v", err)
	}
}
