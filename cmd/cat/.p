I wanted to add syntax highlighting to `cat`, but I didn't like the end result. Please get this closer to the original `cat` from `u-root` use both "flag" and "uflag" (to support flag files).

Behavior should be:
1. -x : enable syntax highlighting
2. If a flag file indicates -x : enable syntax highlighting unless the output of cat is being piped to another program or captured
3. The behavior of rule 2 can be overriden if the user passes `-x` as a cli flag

```ugly, non suckless, non pretty, bloated, code. Incorrectly using chroma too. It does compile and work but it doesn't work like it should, it does however highlight.
package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"

    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
    "github.com/u-root/u-root/pkg/uflag"
)

var (
    syntaxHighlighting = false
)

func cat(reader io.Reader, writer io.Writer) error {
    if _, err := io.Copy(writer, reader); err!= nil {
        return err
    }
    return nil
}

func run(stdin io.Reader, stdout io.Writer, args...string) error {
    if len(args) == 0 {
        if err := cat(stdin, stdout); err!= nil {
            return err
        }
    }
    for _, file := range args {
        if file == "-" {
            err := cat(stdin, stdout)
            if err!= nil {
                return err
            }
            continue
        }
        f, err := os.Open(file)
        if err!= nil {
            return err
        }
        if err := cat(f, stdout); err!= nil {
            return fmt.Errorf("failed to concatenate file %s to given writer", f.Name())
        }
        f.Close()
    }
    return nil
}

func highlightCat(reader io.Reader, writer io.Writer, fileName string) error {
    scanner := bufio.NewScanner(reader)
    lexer := lexers.Match(fileName)
    if lexer == nil {
        lexer = lexers.Fallback
    }
    lexer = chroma.Coalesce(lexer)

    style := styles.Get(os.Getenv("A_COLOR_SCHEME"))
    if style == nil {
        style = styles.Fallback
    }

    formatter := formatters.Get("terminal256")
    if formatter == nil {
        formatter = formatters.Fallback
    }

    for scanner.Scan() {
        iterator, err := lexer.Tokenise(nil, scanner.Text())
        if err!= nil {
            return err
        }
        err = formatter.Format(writer, style, iterator)
        if err!= nil {
            return err
        }
        fmt.Fprintf(writer, "\n")
    }
    return scanner.Err()
}

func main() {
    // Read flags from /etc/cat.flags
    if contents, err := os.ReadFile("/etc/cat.flags"); err == nil {
        args := uflag.FileToArgv(string(contents))
        if len(args) > 0 && args[0] == "-x" {
            syntaxHighlighting = true
        }
    }

    // Read flags from command line
    if len(os.Args) > 1 && os.Args[1] == "-x" {
        syntaxHighlighting = true
    }

    if syntaxHighlighting ||!isPipe() {
        for _, file := range os.Args[1:] {
            if file == "-" {
                if err := highlightCat(os.Stdin, os.Stdout, "stdin"); err!= nil {
                    log.Fatalf("cat failed with: %v", err)
                }
                continue
            }
            f, err := os.Open(file)
            if err!= nil {
                log.Fatalf("cat failed with: %v", err)
            }
            if err := highlightCat(f, os.Stdout, file); err!= nil {
                log.Fatalf("cat failed with: %v", err)
            }
            f.Close()
        }
    } else {
        if err := run(os.Stdin, os.Stdout, os.Args[1:]...); err!= nil {
            log.Fatalf("cat failed with: %v", err)
        }
    }
}

func isPipe() bool {
    fi, _ := os.Stdin.Stat()
    return fi.Mode()&os.ModeNamedPipe!= 0
}
```

```pretty, short, suckless, cat implementation of u-root
package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"

    "github.com/alecthomas/chroma/v2"
    "github.com/alecthomas/chroma/v2/formatters"
    "github.com/alecthomas/chroma/v2/lexers"
    "github.com/alecthomas/chroma/v2/styles"
    "github.com/u-root/u-root/pkg/uflag"
)

var (
    syntaxHighlighting = false
)

func cat(reader io.Reader, writer io.Writer) error {
    if _, err := io.Copy(writer, reader); err!= nil {
        return err
    }
    return nil
}

func run(stdin io.Reader, stdout io.Writer, args...string) error {
    if len(args) == 0 {
        if err := cat(stdin, stdout); err!= nil {
            return err
        }
    }
    for _, file := range args {
        if file == "-" {
            err := cat(stdin, stdout)
            if err!= nil {
                return err
            }
            continue
        }
        f, err := os.Open(file)
        if err!= nil {
            return err
        }
        if err := cat(f, stdout); err!= nil {
            return fmt.Errorf("failed to concatenate file %s to given writer", f.Name())
        }
        f.Close()
    }
    return nil
}

func highlightCat(reader io.Reader, writer io.Writer, fileName string) error {
    scanner := bufio.NewScanner(reader)
    lexer := lexers.Match(fileName)
    if lexer == nil {
        lexer = lexers.Fallback
    }
    lexer = chroma.Coalesce(lexer)

    style := styles.Get(os.Getenv("A_COLOR_SCHEME"))
    if style == nil {
        style = styles.Fallback
    }

    formatter := formatters.Get("terminal256")
    if formatter == nil {
        formatter = formatters.Fallback
    }

    for scanner.Scan() {
        iterator, err := lexer.Tokenise(nil, scanner.Text())
        if err!= nil {
            return err
        }
        err = formatter.Format(writer, style, iterator)
        if err!= nil {
            return err
        }
        fmt.Fprintf(writer, "\n")
    }
    return scanner.Err()
}

func main() {
    // Read flags from /etc/cat.flags
    if contents, err := os.ReadFile("/etc/cat.flags"); err == nil {
        args := uflag.FileToArgv(string(contents))
        if len(args) > 0 && args[0] == "-x" {
            syntaxHighlighting = true
        }
    }

    // Read flags from command line
    if len(os.Args) > 1 && os.Args[1] == "-x" {
        syntaxHighlighting = true
    }

    if syntaxHighlighting ||!isPipe() {
        for _, file := range os.Args[1:] {
            if file == "-" {
                if err := highlightCat(os.Stdin, os.Stdout, "stdin"); err!= nil {
                    log.Fatalf("cat failed with: %v", err)
                }
                continue
            }
            f, err := os.Open(file)
            if err!= nil {
                log.Fatalf("cat failed with: %v", err)
            }
            if err := highlightCat(f, os.Stdout, file); err!= nil {
                log.Fatalf("cat failed with: %v", err)
            }
            f.Close()
        }
    } else {
        if err := run(os.Stdin, os.Stdout, os.Args[1:]...); err!= nil {
            log.Fatalf("cat failed with: %v", err)
        }
    }
}

func isPipe() bool {
    fi, _ := os.Stdin.Stat()
    return fi.Mode()&os.ModeNamedPipe!= 0
}
```

```docs of chroma
Using the library

This is version 2 of Chroma, use the import path:

import "github.com/alecthomas/chroma/v2"

Chroma, like Pygments, has the concepts of lexers, formatters and styles.

Lexers convert source text into a stream of tokens, styles specify how token types are mapped to colours, and formatters convert tokens and styles into formatted output.

A package exists for each of these, containing a global Registry variable with all of the registered implementations. There are also helper functions for using the registry in each package, such as looking up lexers by name or matching filenames, etc.

In all cases, if a lexer, formatter or style can not be determined, nil will be returned. In this situation you may want to default to the Fallback value in each respective package, which provides sane defaults.
Quick start

A convenience function exists that can be used to simply format some source text, without any effort:

err := quick.Highlight(os.Stdout, someSourceCode, "go", "html", "monokai")

Identifying the language

To highlight code, you'll first have to identify what language the code is written in. There are three primary ways to do that:

    Detect the language from its filename.

    lexer := lexers.Match("foo.go")

Explicitly specify the language by its Chroma syntax ID (a full list is available from lexers.Names()).

lexer := lexers.Get("go")

Detect the language from its content.

lexer := lexers.Analyse("package main\n\nfunc main()\n{\n}\n")

In all cases, nil will be returned if the language can not be identified.

if lexer == nil {
  lexer = lexers.Fallback
}

At this point, it should be noted that some lexers can be extremely chatty. To mitigate this, you can use the coalescing lexer to coalesce runs of identical token types into a single token:

lexer = chroma.Coalesce(lexer)

Formatting the output

Once a language is identified you will need to pick a formatter and a style (theme).

style := styles.Get("swapoff")
if style == nil {
  style = styles.Fallback
}
formatter := formatters.Get("html")
if formatter == nil {
  formatter = formatters.Fallback
}

Then obtain an iterator over the tokens:

contents, err := ioutil.ReadAll(r)
iterator, err := lexer.Tokenise(nil, string(contents))

And finally, format the tokens from the iterator:

err := formatter.Format(w, style, iterator)


```