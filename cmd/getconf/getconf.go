package main

import (
	"flag"
	"fmt"
	"os"

	"a-utils/pkg/ccmd"
	"github.com/tklauser/go-sysconf"
)

var sysconfVars = map[string]int{
	"CLK_TCK":                      sysconf.SC_CLK_TCK,
	"ARG_MAX":                      sysconf.SC_ARG_MAX,
	"PAGE_SIZE":                    sysconf.SC_PAGE_SIZE,
	"AIO_LISTIO_MAX":               sysconf.SC_AIO_LISTIO_MAX,
	"AIO_MAX":                      sysconf.SC_AIO_MAX,
	"AIO_PRIO_DELTA_MAX":           sysconf.SC_AIO_PRIO_DELTA_MAX,
	"ATEXIT_MAX":                   sysconf.SC_ATEXIT_MAX,
	"BC_BASE_MAX":                  sysconf.SC_BC_BASE_MAX,
	"BC_DIM_MAX":                   sysconf.SC_BC_DIM_MAX,
	"BC_SCALE_MAX":                 sysconf.SC_BC_SCALE_MAX,
	"BC_STRING_MAX":                sysconf.SC_BC_STRING_MAX,
	"CHILD_MAX":                    sysconf.SC_CHILD_MAX,
	"COLL_WEIGHTS_MAX":             sysconf.SC_COLL_WEIGHTS_MAX,
	"DELAYTIMER_MAX":               sysconf.SC_DELAYTIMER_MAX,
	"EXPR_NEST_MAX":                sysconf.SC_EXPR_NEST_MAX,
	"GETGR_R_SIZE_MAX":             sysconf.SC_GETGR_R_SIZE_MAX,
	"GETPW_R_SIZE_MAX":             sysconf.SC_GETPW_R_SIZE_MAX,
	"HOST_NAME_MAX":                sysconf.SC_HOST_NAME_MAX,
	"IOV_MAX":                      sysconf.SC_IOV_MAX,
	"LINE_MAX":                     sysconf.SC_LINE_MAX,
	"LOGIN_NAME_MAX":               sysconf.SC_LOGIN_NAME_MAX,
	"MQ_OPEN_MAX":                  sysconf.SC_MQ_OPEN_MAX,
	"MQ_PRIO_MAX":                  sysconf.SC_MQ_PRIO_MAX,
	"NGROUPS_MAX":                  sysconf.SC_NGROUPS_MAX,
	"OPEN_MAX":                     sysconf.SC_OPEN_MAX,
	"THREAD_DESTRUCTOR_ITERATIONS": sysconf.SC_THREAD_DESTRUCTOR_ITERATIONS,
	"THREAD_KEYS_MAX":              sysconf.SC_THREAD_KEYS_MAX,
	"THREAD_STACK_MIN":             sysconf.SC_THREAD_STACK_MIN,
	"THREAD_THREADS_MAX":           sysconf.SC_THREAD_THREADS_MAX,
	"RE_DUP_MAX":                   sysconf.SC_RE_DUP_MAX,
	"RTSIG_MAX":                    sysconf.SC_RTSIG_MAX,
	"SEM_NSEMS_MAX":                sysconf.SC_SEM_NSEMS_MAX,
	"SEM_VALUE_MAX":                sysconf.SC_SEM_VALUE_MAX,
	"SIGQUEUE_MAX":                 sysconf.SC_SIGQUEUE_MAX,
	"STREAM_MAX":                   sysconf.SC_STREAM_MAX,
	"SYMLOOP_MAX":                  sysconf.SC_SYMLOOP_MAX,
	"TIMER_MAX":                    sysconf.SC_TIMER_MAX,
	"TTY_NAME_MAX":                 sysconf.SC_TTY_NAME_MAX,
	"TZNAME_MAX":                   sysconf.SC_TZNAME_MAX,

	"ADVISORY_INFO":         sysconf.SC_ADVISORY_INFO,
	"ASYNCHRONOUS_IO":       sysconf.SC_ASYNCHRONOUS_IO,
	"BARRIERS":              sysconf.SC_BARRIERS,
	"CLOCK_SELECTION":       sysconf.SC_CLOCK_SELECTION,
	"CPUTIME":               sysconf.SC_CPUTIME,
	"FSYNC":                 sysconf.SC_FSYNC,
	"IPV6":                  sysconf.SC_IPV6,
	"JOB_CONTROL":           sysconf.SC_JOB_CONTROL,
	"MAPPED_FILES":          sysconf.SC_MAPPED_FILES,
	"MEMLOCK":               sysconf.SC_MEMLOCK,
	"MEMLOCK_RANGE":         sysconf.SC_MEMLOCK_RANGE,
	"MEMORY_PROTECTION":     sysconf.SC_MEMORY_PROTECTION,
	"MESSAGE_PASSING":       sysconf.SC_MESSAGE_PASSING,
	"MONOTONIC_CLOCK":       sysconf.SC_MONOTONIC_CLOCK,
	"PRIORITIZED_IO":        sysconf.SC_PRIORITIZED_IO,
	"PRIORITY_SCHEDULING":   sysconf.SC_PRIORITY_SCHEDULING,
	"RAW_SOCKETS":           sysconf.SC_RAW_SOCKETS,
	"READER_WRITER_LOCKS":   sysconf.SC_READER_WRITER_LOCKS,
	"REALTIME_SIGNALS":      sysconf.SC_REALTIME_SIGNALS,
	"REGEXP":                sysconf.SC_REGEXP,
	"SAVED_IDS":             sysconf.SC_SAVED_IDS,
	"SEMAPHORES":            sysconf.SC_SEMAPHORES,
	"SHARED_MEMORY_OBJECTS": sysconf.SC_SHARED_MEMORY_OBJECTS,
	"SHELL":                 sysconf.SC_SHELL,
	"SPAWN":                 sysconf.SC_SPAWN,
	"SPIN_LOCKS":            sysconf.SC_SPIN_LOCKS,
	"SPORADIC_SERVER":       sysconf.SC_SPORADIC_SERVER,
	//"SS_REPL_MAX":                sysconf.SC_SS_REPL_MAX,
	"SYNCHRONIZED_IO":            sysconf.SC_SYNCHRONIZED_IO,
	"THREAD_ATTR_STACKADDR":      sysconf.SC_THREAD_ATTR_STACKADDR,
	"THREAD_ATTR_STACKSIZE":      sysconf.SC_THREAD_ATTR_STACKSIZE,
	"THREAD_CPUTIME":             sysconf.SC_THREAD_CPUTIME,
	"THREAD_PRIO_INHERIT":        sysconf.SC_THREAD_PRIO_INHERIT,
	"THREAD_PRIO_PROTECT":        sysconf.SC_THREAD_PRIO_PROTECT,
	"THREAD_PRIORITY_SCHEDULING": sysconf.SC_THREAD_PRIORITY_SCHEDULING,
	"THREAD_PROCESS_SHARED":      sysconf.SC_THREAD_PROCESS_SHARED,
	//"THREAD_ROBUST_PRIO_INHERIT": sysconf.SC_THREAD_ROBUST_PRIO_INHERIT,
	//"THREAD_ROBUST_PRIO_PROTECT": sysconf.SC_THREAD_ROBUST_PRIO_PROTECT,
	"THREAD_SAFE_FUNCTIONS":  sysconf.SC_THREAD_SAFE_FUNCTIONS,
	"THREAD_SPORADIC_SERVER": sysconf.SC_THREAD_SPORADIC_SERVER,
	"THREADS":                sysconf.SC_THREADS,
	"TIMEOUTS":               sysconf.SC_TIMEOUTS,
	"TIMERS":                 sysconf.SC_TIMERS,
	"TRACE":                  sysconf.SC_TRACE,
	"TRACE_EVENT_FILTER":     sysconf.SC_TRACE_EVENT_FILTER,
	"TRACE_EVENT_NAME_MAX":   sysconf.SC_TRACE_EVENT_NAME_MAX,
	"TRACE_INHERIT":          sysconf.SC_TRACE_INHERIT,
	"TRACE_LOG":              sysconf.SC_TRACE_LOG,
	"TRACE_NAME_MAX":         sysconf.SC_TRACE_NAME_MAX,
	"TRACE_SYS_MAX":          sysconf.SC_TRACE_SYS_MAX,
	"TRACE_USER_EVENT_MAX":   sysconf.SC_TRACE_USER_EVENT_MAX,
	"TYPED_MEMORY_OBJECTS":   sysconf.SC_TYPED_MEMORY_OBJECTS,
	"VERSION":                sysconf.SC_VERSION,

	"V7_ILP32_OFF32":  sysconf.SC_V7_ILP32_OFF32,
	"V7_ILP32_OFFBIG": sysconf.SC_V7_ILP32_OFFBIG,
	"V7_LP64_OFF64":   sysconf.SC_V7_LP64_OFF64,
	"V7_LPBIG_OFFBIG": sysconf.SC_V7_LPBIG_OFFBIG,

	"V6_ILP32_OFF32":  sysconf.SC_V6_ILP32_OFF32,
	"V6_ILP32_OFFBIG": sysconf.SC_V6_ILP32_OFFBIG,
	"V6_LP64_OFF64":   sysconf.SC_V6_LP64_OFF64,
	"V6_LPBIG_OFFBIG": sysconf.SC_V6_LPBIG_OFFBIG,

	"2_C_BIND":         sysconf.SC_2_C_BIND,
	"2_C_DEV":          sysconf.SC_2_C_DEV,
	"2_C_VERSION":      sysconf.SC_2_C_VERSION,
	"2_CHAR_TERM":      sysconf.SC_2_CHAR_TERM,
	"2_FORT_DEV":       sysconf.SC_2_FORT_DEV,
	"2_FORT_RUN":       sysconf.SC_2_FORT_RUN,
	"2_LOCALEDEF":      sysconf.SC_2_LOCALEDEF,
	"2_PBS":            sysconf.SC_2_PBS,
	"2_PBS_ACCOUNTING": sysconf.SC_2_PBS_ACCOUNTING,
	"2_PBS_CHECKPOINT": sysconf.SC_2_PBS_CHECKPOINT,
	"2_PBS_LOCATE":     sysconf.SC_2_PBS_LOCATE,
	"2_PBS_MESSAGE":    sysconf.SC_2_PBS_MESSAGE,
	"2_PBS_TRACK":      sysconf.SC_2_PBS_TRACK,
	"2_SW_DEV":         sysconf.SC_2_SW_DEV,
	"2_UPE":            sysconf.SC_2_UPE,
	"2_VERSION":        sysconf.SC_2_VERSION,

	"XOPEN_CRYPT":            sysconf.SC_XOPEN_CRYPT,
	"XOPEN_ENH_I18N":         sysconf.SC_XOPEN_ENH_I18N,
	"XOPEN_REALTIME":         sysconf.SC_XOPEN_REALTIME,
	"XOPEN_REALTIME_THREADS": sysconf.SC_XOPEN_REALTIME_THREADS,
	"XOPEN_SHM":              sysconf.SC_XOPEN_SHM,
	"XOPEN_STREAMS":          sysconf.SC_XOPEN_STREAMS,
	"XOPEN_UNIX":             sysconf.SC_XOPEN_UNIX,
	"XOPEN_VERSION":          sysconf.SC_XOPEN_VERSION,
	"XOPEN_XCU_VERSION":      sysconf.SC_XOPEN_XCU_VERSION,

	"PHYS_PAGES":       sysconf.SC_PHYS_PAGES,
	"AVPHYS_PAGES":     sysconf.SC_AVPHYS_PAGES,
	"NPROCESSORS_CONF": sysconf.SC_NPROCESSORS_CONF,
	"NPROCESSORS_ONLN": sysconf.SC_NPROCESSORS_ONLN,
	"UIO_MAXIOV":       sysconf.SC_UIO_MAXIOV,
}

func printSysconfValue(name string) {
	scConst, found := sysconfVars[name]
	if !found {
		fmt.Fprintf(os.Stderr, "Unknown variable: %s\n", name)
		os.Exit(1)
	}

	value, err := sysconf.Sysconf(scConst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting sysconf value: %v\n", err)
		os.Exit(1)
	}
	fmt.Println(value)
}

func printAllSysconf() {
	for name := range sysconfVars {
		fmt.Printf("%s: ", name)
		printSysconfValue(name)
	}
}

func main() {
	version := flag.Bool("v", false, "Show the version of the sysconf variable")
	all := flag.Bool("a", false, "Print all sysconf values")

	cmdInfo := &ccmd.CmdInfo{
		Name:        "getconf",
		Authors:     []string{"yourname"},
		Description: "Get system configuration values",
		Synopsis:    "<|-v|-a|> var [path]",
		Behavior:    "Prints sysconf values or handles the specified path.",
	}

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

	if *version {
		// Print specific sysconf variable value
		if len(args) != 1 {
			fmt.Println("Usage: getconf -v spec")
			os.Exit(1)
		}
		printSysconfValue(args[0])
		return
	}

	if *all {
		// Print all sysconf values
		printAllSysconf()
		return
	}

	if len(args) == 1 {
		// Print value for a single sysconf variable
		printSysconfValue(args[0])
		return
	}

	if len(args) == 2 {
		// Handle path variable, not implemented in detail
		fmt.Printf("Path variable: %s\n", args[1])
		// Add path variable handling if needed
		return
	}
}
