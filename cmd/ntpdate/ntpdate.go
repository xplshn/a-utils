package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/xplshn/a-utils/pkg/ccmd"
	"github.com/u-root/u-root/pkg/ntpdate"
)

const fallback = "pool.ntp.org"

func main() {
	cmdInfo := &ccmd.CmdInfo{
		Name:        "ntpdate",
		Authors:     []string{"u-root Authors", "xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "queries NTP server(s) for time and updates system time.",
		Synopsis:    "<options> <server domain>",
		CustomFields: map[string]interface{}{
			"1_Examples": `# Update time and set hardware clock:
  \$ ntpdate --rtc

# Use specific NTP server:
  \$ ntpdate pool.ntp.org`,
			"2_Behavior": `ntpdate queries the provided NTP server(s) and adjusts the system time.
If --rtc is specified, it updates the hardware clock (RTC) as well.
By default, the servers are read from /etc/ntp.conf. If servers are passed on
the command line, those are tried first. As a fallback, pool.ntp.org is used.`,
		},
	}

	config := flag.String("config", ntpdate.DefaultNTPConfig, "NTP config file")
	setRTC := flag.Bool("rtc", false, "Set RTC time as well")
	verbose := flag.Bool("verbose", false, "Verbose output")

	helpPage, err := cmdInfo.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}

	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()

	if len(flag.Args()) == 0 {
		fmt.Print(helpPage)
		os.Exit(0)
	}

	if *verbose {
		ntpdate.Debug = log.Printf
	}

	server, offset, err := ntpdate.SetTime(flag.Args(), *config, fallback, *setRTC)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	plus := ""
	if offset > 0 {
		plus = "+"
	}
	log.Printf("adjust time server %s offset %s%f sec", server, plus, offset)
}
