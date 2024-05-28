package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func concatenateEnvVars(suffix string) string {
	var result []string
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if strings.HasSuffix(parts[0], suffix) {
			result = append(result, parts[1])
		}
	}
	return strings.Join(result, ":")
}

func main() {
	PELF_BINDIRS := concatenateEnvVars("_bindir")
	PELF_LIBDIRS := concatenateEnvVars("_libs")

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--export" {
		os.Setenv("PELF_BINDIRS", PELF_BINDIRS)
		os.Setenv("PELF_LIBDIRS", PELF_LIBDIRS)
	} else if len(args) > 0 {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Env = append(os.Environ(), fmt.Sprintf("LD_LIBRARY_PATH=%s", PELF_LIBDIRS), fmt.Sprintf("PATH=%s:%s", os.Getenv("PATH"), PELF_BINDIRS))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error executing command: %v\n", err)
			os.Exit(1)
		}
	} else {
		fmt.Printf("PELF_BINDIRS=\"%s\"\n", PELF_BINDIRS)
		fmt.Printf("PELF_LIBDIRS=\"%s\"\n", PELF_LIBDIRS)
	}
}
