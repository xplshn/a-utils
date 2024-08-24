// Copyright (c) 2024-2024 xplshn                       [3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/formatters"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/alecthomas/chroma/v2/styles"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "ccat",
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Concatenates files and prints them to stdout with Syntax Highlighting",
		Synopsis:    "<|--styles|--style [SYTHX_FILE]|> [FILE/s]",
		CustomFields: map[string]interface{}{
			"1_Behavior": "If no files are specified, read from stdin.",
			"2_Notes":    "The following env variables allow you to set the Style and Formatter to be used:\n  A_SYHX_COLOR_SCHEME: string: Acceptable values include any of the lines that `--styles` outputs\n  A_SYHX_FORMATTER: string: Acceptable values include: terminal8, terminal16 and terminal256\ns",
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

	stylesFlag := flag.Bool("styles", false, "List available styles")
	styleFile := flag.String("style", "", "Load custom style from file")
	flag.Parse()

	if *stylesFlag {
		fmt.Println("Available styles:")
		for _, style := range styles.Names() {
			fmt.Println(style)
		}
		return
	}

	var customStyleName string
	if envStyleFile := os.Getenv("A_SYHX_CUSTOM_COLOR_SCHEME"); envStyleFile != "" {
		*styleFile = envStyleFile
	}
	if *styleFile != "" {
		var err error
		customStyleName, err = loadCustomStyle(*styleFile)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading custom style:", err)
			os.Exit(1)
		}
	}

	args := flag.Args()
	if err := run(os.Stdin, os.Stdout, customStyleName, args...); err != nil {
		fmt.Fprintln(os.Stderr, "cat failed:", err)
		os.Exit(1)
	}
}

func loadCustomStyle(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open style file %s: %v", filePath, err)
	}
	defer file.Close()

	style, err := chroma.NewXMLStyle(file)
	if err != nil {
		return "", fmt.Errorf("failed to parse style file %s: %v", filePath, err)
	}

	styles.Register(style)
	fmt.Printf("Loaded custom style: %s\n", style.Name)
	return style.Name, nil
}

func run(stdin io.Reader, stdout io.Writer, customStyleName string, args ...string) error {
	if len(args) == 0 {
		return highlightCat(stdin, stdout, "stdin", customStyleName)
	}

	for _, file := range args {
		var reader io.Reader
		if file == "-" {
			reader = stdin
		} else {
			f, err := os.Open(file)
			if err != nil {
				return fmt.Errorf("failed to open file %s: %v", file, err)
			}
			defer f.Close()
			reader = f
		}
		if err := highlightCat(reader, stdout, file, customStyleName); err != nil {
			return err
		}
	}
	return nil
}

func highlightCat(reader io.Reader, writer io.Writer, fileName string, customStyleName string) error {
	contents, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Sanitize input
	sanitizedContents := sanitizeInput(string(contents))

	// Detect the language from content
	lexer := lexers.Analyse(sanitizedContents)
	if lexer == nil {
		lexer = lexers.Fallback
	}
	lexer = chroma.Coalesce(lexer)

	var style *chroma.Style
	if customStyleName != "" {
		style = styles.Get(customStyleName)
	} else {
		style = styles.Get(os.Getenv("A_SYHX_COLOR_SCHEME"))
		if style == nil {
			style = styles.Fallback
		}
	}

	formatterName := os.Getenv("A_SYHX_FORMATTER")
	if formatterName == "" {
		formatterName = "terminal16"
	}

	formatter := formatters.Get(formatterName)
	if formatter == nil {
		formatter = formatters.Fallback
	}

	iterator, err := lexer.Tokenise(nil, string(contents))
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	if err := formatter.Format(&buf, style, iterator); err != nil {
		return err
	}

	_, err = io.Copy(writer, &buf)
	return err
}

func sanitizeInput(input string) string {
	var sanitized strings.Builder
	for _, r := range input {
		if unicode.IsPrint(r) && !unicode.IsControl(r) {
			sanitized.WriteRune(r)
		}
	}
	return sanitized.String()
}
