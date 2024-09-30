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
	"time"

	"github.com/xplshn/a-utils/pkg/ccmd"
)

const (
	Prefix = "dial: "
	Debug  = false
)

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "dial",
		Authors:     []string{"as"},
		Repository:  "https://github.com/as/dial",
		Description: "Dial a network endpoint and optionally run a command for each new connection.",
		Synopsis:    "dial [options] host:port [cmd ...]",
		CustomFields: map[string]interface{}{
			"1_Examples": `Speak HTTP:
  \$ echo GET / HTTP/1.1 | dial example.com:80

RDP tunnel through port 80:
  \$ listen :80 dial 10.2.64.20:3389`,
			"2_Behavior": `Dial establishes a connection with the listener on the
remote host and runs cmd. Cmd's three standard file
descriptors (stdin, stdout+stderr) are connected to the
listener via proto (default tcp).

If cmd is not given, the standard file descriptors are
instead connected to dial's standard input, output, and
error.`,
		},
	}

	showHelp := flag.Bool("h", false, "Show help")
	verbose := flag.Bool("v", false, "Verbose output")
	keepAlive := flag.Bool("k", false, "Enable TCP keep-alive")
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

	conn, err := net.Dial(*protocol, server)
	if err != nil {
		handleFatalError(err)
	}

	logVerbose(*verbose, "connected to:", conn.RemoteAddr())

	// Enable TCP keep-alive if the flag is set
	if tcpConn, ok := conn.(*net.TCPConn); ok && *keepAlive {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(3 * time.Minute) // Adjust the keep-alive period as needed
	}

	if *muxMode {
		handleMultiplexedStream(conn, *activeLimit, *keepAlive, command...)
	} else {
		handleStream(conn, *activeLimit, *keepAlive, command...)
	}
}

func handleStream(conn net.Conn, limit int, keepAlive bool, command ...string) {
	if len(command) == 0 {
		err := handleTerminalIO(conn)
		printError(err)
	} else {
		err := executeCommand(conn, command[0], command[1:]...)
		printError(err)
	}
}

func handleMultiplexedStream(conn net.Conn, limit int, keepAlive bool, command ...string) {
	concurrencyLimit := make(chan bool, limit)
	connChan := make(chan io.ReadWriter)

	go multiplexer(connChan)

	for {
		concurrencyLimit <- true
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
