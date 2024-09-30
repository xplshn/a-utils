// Copyright (c) as 2016, 2024-2024 xplshn				[3BSD]
// For more details refer to https://github.com/xplshn/a-utils
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	Prefix = "listen: "
	Debug  = false
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "listen",
		Authors:     []string{"as", "xplshn"},
		Repository:  "https://github.com/as/listen",
		Description: "Listen on a network port and optionally run a command for each new connection.",
		Synopsis:    "listen [options] port [cmd ...]",
		CustomFields: map[string]interface{}{
			"Examples": `Serve index.html over HTTP:
  \$ listen :80 cat index.html

Forward connections to google.com:
  \$ listen :80 dial google.com:80`,
		},
	}

	showHelp := flag.Bool("h", false, "Show help")
	verbose := flag.Bool("v", false, "Verbose output")
	keepAlive := flag.Int("k", 0, "Terminate after n calls")
	muxMode := flag.Bool("m", false, "Mux: broadcast traffic to all clients")
	activeLimit := flag.Int("a", 4096, "Active connection limit")
	protocol := flag.String("n", "tcp4", "Network protocol")

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()

	if *showHelp || len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	server := flag.Args()[0]
	command := flag.Args()[1:]

	listener, err := net.Listen(*protocol, server)
	if err != nil {
		handleFatalError(err)
	}

	logVerbose(*verbose, "announce:", listener.Addr())

	if *muxMode {
		handleMultiplexedStream(listener, *activeLimit, *keepAlive, command...)
	} else {
		handleStream(listener, *activeLimit, *keepAlive, command...)
	}
}

func handleStream(listener net.Listener, limit, keepAlive int, command ...string) {
	concurrencyLimit := make(chan bool, limit)
	callCount := 0

	for {
		concurrencyLimit <- true
		conn, err := listener.Accept()
		if err != nil {
			printError(err)
			<-concurrencyLimit
			continue
		}

		go func() {
			defer func() { <-concurrencyLimit }()
			defer conn.Close()

			if len(command) == 0 {
				err := handleTerminalIO(conn)
				printError(err)
			} else {
				err := executeCommand(conn, command[0], command[1:]...)
				printError(err)
			}

			// Increment and check keepAlive condition
			if keepAlive > 0 {
				callCount++
				if callCount >= keepAlive {
					fmt.Println("Terminating after", keepAlive, "calls")
					os.Exit(0)
				}
			}
		}()
	}
}

func handleMultiplexedStream(listener net.Listener, limit, keepAlive int, command ...string) {
	concurrencyLimit := make(chan bool, limit)
	connChan := make(chan io.ReadWriter)
	callCount := 0

	go multiplexer(connChan)

	for {
		concurrencyLimit <- true
		conn, err := listener.Accept()
		if err != nil {
			printError(err)
			<-concurrencyLimit
			continue
		}

		go func() {
			defer func() { <-concurrencyLimit }()
			defer conn.Close()

			connChan <- conn
			muxedConn := <-connChan

			if len(command) == 0 {
				err := handleTerminalIO(muxedConn.(net.Conn))
				printError(err)
			} else {
				err := executeCommand(muxedConn, command[0], command[1:]...)
				printError(err)
			}

			// Increment and check keepAlive condition
			if keepAlive > 0 {
				callCount++
				if callCount >= keepAlive {
					fmt.Println("Terminating after", keepAlive, "calls")
					os.Exit(0)
				}
			}
		}()
	}
}

func multiplexer(connChan chan io.ReadWriter) {
	writers := []io.Writer{}
	readers := []io.Reader{}

	for conn := range connChan {
		writers = append(writers, conn)
		readers = append(readers, conn)

		multiReader := io.MultiReader(readers...)
		multiWriter := io.MultiWriter(writers...)

		connChan <- bufio.NewReadWriter(bufio.NewReader(multiReader), bufio.NewWriter(multiWriter))
	}
}

func handleTerminalIO(conn net.Conn) error {
	finishChan := make(chan error)
	defer close(finishChan)

	go func() {
		_, err := io.Copy(os.Stdout, conn)
		finishChan <- err
	}()

	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			printError("stdin|net", err)
		}
	}()

	return <-finishChan
}

func executeCommand(rw io.ReadWriter, cmd string, args ...string) error {
	cmdExec := exec.Command(cmd, args...)
	stdin, err := cmdExec.StdinPipe()
	if err != nil {
		return err
	}
	pipeReader, pipeWriter := io.Pipe()
	defer pipeWriter.Close()

	cmdExec.Stdout, cmdExec.Stderr = pipeWriter, pipeWriter

	if err := cmdExec.Start(); err != nil {
		return err
	}

	finishChan := make(chan error)
	defer close(finishChan)

	go func() {
		_, err := io.Copy(stdin, rw)
		pipeWriter.Close()
		finishChan <- err
	}()

	go func() {
		_, err := io.Copy(rw, pipeReader)
		if err != nil {
			printError("command|net", err)
		}
	}()

	if err := <-finishChan; err != nil {
		printError("net|command", err)
	}

	stdin.Close()
	return cmdExec.Wait()
}

func handleFatalError(err error) {
	if err != nil {
		printError(err)
		os.Exit(1)
	}
}

func logVerbose(verbose bool, args ...interface{}) {
	if verbose {
		fmt.Fprintln(os.Stderr, args...)
	}
}

func printError(args ...interface{}) {
	fmt.Fprint(os.Stderr, Prefix)
	fmt.Fprintln(os.Stderr, args...)
}
