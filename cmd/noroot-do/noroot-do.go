package main

import (
	"embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/xattr"
	"github.com/xplshn/a-utils/pkg/ccmd"
)

//go:embed bwrap
var embeddedBwrap embed.FS

const (
	xattrKeyRootFS      = "user.SetRootfs"
	xattrKeyBwrapToggle = "user.UseEmbeddedBwrap"
	xattrKeyModeFlags   = "user.ModeFlags"
	xattrKeySediment    = "user.Sediment"
)

// Save the root filesystem path using xattr on the executable itself (os.Args[0])
func saveRootFSConfig(path string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to determine absolute path of rootfs: %v", err)
	}
	if err := xattr.Set(executable, xattrKeyRootFS, []byte(absPath)); err != nil {
		return fmt.Errorf("failed to set xattr for rootfs: %v", err)
	}
	return nil
}

// Load the root filesystem path from xattr
func loadRootFSConfig() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to determine current executable: %v", err)
	}
	rootfs, err := xattr.Get(executable, xattrKeyRootFS)
	if err != nil {
		return "", fmt.Errorf("failed to get xattr for rootfs: %v", err)
	}
	return string(rootfs), nil
}

// Remove the root filesystem xattr
func removeRootFSConfig() error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	return xattr.Remove(executable, xattrKeyRootFS)
}

// Save bwrap toggle using xattr
func saveBwrapToggle(useEmbedded bool) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	value := "0"
	if useEmbedded {
		value = "1"
	}
	if err := xattr.Set(executable, xattrKeyBwrapToggle, []byte(value)); err != nil {
		return fmt.Errorf("failed to set xattr for bwrap toggle: %v", err)
	}
	return nil
}

// Load bwrap toggle from xattr
func loadBwrapToggle() (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to determine current executable: %v", err)
	}
	value, err := xattr.Get(executable, xattrKeyBwrapToggle)
	if err != nil {
		return false, nil // default to system bwrap if no xattr set
	}
	return string(value) == "1", nil
}

// Check if the root filesystem is set by checking the xattr
func checkRootFSSet() bool {
	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to determine current executable: %v", err)
	}
	_, err = xattr.Get(executable, xattrKeyRootFS)
	return err == nil
}

// Check if the embedded bwrap should be used
func checkUseEmbeddedBwrap() bool {
	useEmbedded, err := loadBwrapToggle()
	if err != nil {
		log.Printf("Error loading bwrap toggle: %v, using system bwrap by default", err)
		return false
	}
	return useEmbedded
}

// Save mode flags using xattr
func saveModeFlags(mode string, flags string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	if err := xattr.Set(executable, fmt.Sprintf("%s.%s", xattrKeyModeFlags, mode), []byte(flags)); err != nil {
		return fmt.Errorf("failed to set xattr for mode flags: %v", err)
	}
	return nil
}

// Load mode flags from xattr
func loadModeFlags(mode string) (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to determine current executable: %v", err)
	}
	flags, err := xattr.Get(executable, fmt.Sprintf("%s.%s", xattrKeyModeFlags, mode))
	if err != nil {
		return "", fmt.Errorf("failed to get xattr for mode flags: %v", err)
	}
	return string(flags), nil
}

// Remove mode flags using xattr
func removeModeFlags(mode string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	return xattr.Remove(executable, fmt.Sprintf("%s.%s", xattrKeyModeFlags, mode))
}

// Save sediment flag using xattr
func saveSedimentFlag(mode string) error {
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine current executable: %v", err)
	}
	if err := xattr.Set(executable, fmt.Sprintf("%s.%s", xattrKeySediment, mode), []byte("1")); err != nil {
		return fmt.Errorf("failed to set xattr for sediment flag: %v", err)
	}
	return nil
}

// Check if a mode is sediment (RO)
func isSediment(mode string) (bool, error) {
	executable, err := os.Executable()
	if err != nil {
		return false, fmt.Errorf("failed to determine current executable: %v", err)
	}
	value, err := xattr.Get(executable, fmt.Sprintf("%s.%s", xattrKeySediment, mode))
	if err != nil {
		return false, nil // default to not sediment if no xattr set
	}
	return string(value) == "1", nil
}

// Run a command inside the chroot environment using bwrap
func runBwrapCommand(rootfs string, mode string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided to run inside the chroot")
	}

	bwrapCmd := "bwrap"

	if checkUseEmbeddedBwrap() {
		tmpFile, err := os.CreateTemp("", "bwrap-embedded-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file for embedded bwrap: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		// Open the embedded bwrap binary from the filesystem
		bwrapBinary, err := embeddedBwrap.Open("bwrap")
		if err != nil {
			return fmt.Errorf("failed to open embedded bwrap binary: %v", err)
		}
		defer bwrapBinary.Close()

		// Copy the content of the embedded bwrap binary to the temporary file
		if _, err := io.Copy(tmpFile, bwrapBinary); err != nil {
			return fmt.Errorf("failed to write embedded bwrap to temp file: %v", err)
		}

		// Close the file after writing to it
		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %v", err)
		}

		// Change the permissions of the file using the file name
		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return fmt.Errorf("failed to make temp bwrap executable: %v", err)
		}

		bwrapCmd = tmpFile.Name()
	}

	// Prepare the bwrap command with default arguments
	bwrapArgs := []string{
		"--bind", rootfs, "/", // Mount the rootfs as '/'
		//"--unshare-all",       // Prevents affecting the host system
		//"--share-net",                                       // Let the rootfs have network access
		//"--ro-bind-try", "/etc/localtime", "/etc/localtime", // Share local timezone to be used
		//"--ro-bind-try", "/etc/hostname", "/etc/hostname", // Teach the rootfs what its surname is
		//"--ro-bind-try", "/etc/resolv.conf", "/etc/resolv.conf", // Guide the rootfs towards its first networked steps
		//"--ro-bind-try", "/etc/passwd", "/etc/passwd", // Teach it about its elders & family
		//"--ro-bind-try", "/etc/group", "/etc/group", // Teach it about the importance of social roles
		//"--ro-bind-try", "/etc/hosts", "/etc/hosts", // Make it befriend the local & remote neighbours
		//"--ro-bind-try", "/etc/nsswitch.conf", "/etc/nsswitch.conf", // No comments, forgot what this file is
	}

	// Add mode-specific arguments
	modeFlags, err := loadModeFlags(mode)
	if err != nil {
		return fmt.Errorf("failed to load mode flags: %v", err)
	}
	bwrapArgs = append(bwrapArgs, strings.Split(modeFlags, " ")...)

	// Append user-supplied arguments (command and its options)
	bwrapArgs = append(bwrapArgs, "--")
	bwrapArgs = append(bwrapArgs, args...)

	cmd := exec.Command(bwrapCmd, bwrapArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

func cmdSet(rootfs string) {
	absRootfs, err := filepath.Abs(rootfs)
	if err != nil {
		log.Fatalf("Error determining absolute path: %v", err)
	}
	if err := saveRootFSConfig(absRootfs); err != nil {
		log.Fatalf("Error saving rootfs config: %v", err)
	}
	fmt.Printf("Root filesystem set to: %s\n", absRootfs)
}

func cmdUnset() {
	if !checkRootFSSet() {
		log.Fatalf("No rootfs is currently set.")
	}
	if err := removeRootFSConfig(); err != nil {
		log.Fatalf("Error removing rootfs config: %v", err)
	}
	fmt.Println("Root filesystem unset successfully.")
}

func cmdInfo() {
	if !checkRootFSSet() {
		fmt.Println("rootfs: unset")
	}
	rootfs, err := loadRootFSConfig()
	if err != nil {
		log.Fatalf("Error loading rootfs config: %v", err)
	}
	fmt.Printf("rootfs: %s\n", rootfs)

	useEmbedded := checkUseEmbeddedBwrap()
	if useEmbedded {
		fmt.Println("bwrap: embedded")
	} else {
		fmt.Println("bwrap: system")
	}

	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to determine current executable: %v", err)
		return
	}

	modes, err := xattr.List(executable)
	if err != nil {
		log.Fatalf("Failed to list xattrs: %v", err)
	}

	for _, mode := range modes {
		if strings.HasPrefix(mode, xattrKeyModeFlags+".") {
			modeName := strings.TrimPrefix(mode, xattrKeyModeFlags+".")
			flags, err := loadModeFlags(modeName)
			if err != nil {
				log.Printf("Error loading flags for mode %s: %v", modeName, err)
				continue
			}
			fmt.Printf("Mode: \x1b[94m%s\x1b[0m, Flags: %s\n", modeName, flags)

			isSediment, err := isSediment(modeName)
			if err != nil {
				log.Printf("Error checking sediment for mode %s: %v", modeName, err)
				continue
			}
			if isSediment {
				fmt.Printf("Mode %s is read-only (sedimented).\n", modeName)
			}
		}
	}
}

func cmdToggleEmbeddedBwrap() {
	useEmbedded := checkUseEmbeddedBwrap()
	if err := saveBwrapToggle(!useEmbedded); err != nil {
		log.Fatalf("Error toggling embedded bwrap: %v", err)
	}
	if !useEmbedded {
		fmt.Println("Now using embedded bwrap.")
	} else {
		fmt.Println("Now using system bwrap.")
	}
}

func cmdRun(mode string, args []string) {
	if !checkRootFSSet() {
		log.Fatalf("No rootfs is currently set.")
	}

	rootfs, err := loadRootFSConfig()
	if err != nil {
		log.Fatalf("Error loading rootfs config: %v", err)
	}

	err = runBwrapCommand(rootfs, mode, args)
	if err != nil {
		log.Fatalf("Error running command in chroot: %v", err)
	}
}

func cmdSetModeFlags(mode string, flags string) {
	isSediment, err := isSediment(mode)
	if err != nil {
		log.Fatalf("Error checking sediment for mode %s: %v", mode, err)
	}
	if isSediment {
		log.Fatalf("Mode %s is read-only (sedimented) and cannot be modified.", mode)
	}
	if err := saveModeFlags(mode, flags); err != nil {
		log.Fatalf("Error saving mode flags: %v", err)
	}
	if err == nil {
		fmt.Printf("Successfully configured mode \"%s\"\n", mode)
	} else {
		fmt.Printf("%v", err)
	}
}

func cmdRemoveMode(mode string) {
	isSediment, err := isSediment(mode)
	if err != nil {
		log.Fatalf("Error checking sediment for mode %s: %v", mode, err)
	}
	if isSediment {
		log.Fatalf("Mode %s is read-only (sedimented) and cannot be removed.", mode)
	}

	if err := removeModeFlags(mode); err != nil {
		log.Fatalf("Error removing mode %s: %v", mode, err)
	}
	fmt.Printf("Mode %s removed successfully.\n", mode)
}

func cmdSetSediment(mode string) {
	if err := saveSedimentFlag(mode); err != nil {
		log.Fatalf("Error setting sediment flag for mode %s: %v", mode, err)
	}
	fmt.Printf("Mode %s set to read-only (sedimented).\n", mode)
}

func main() {
	setFlag := flag.String("set", "", "Set the root filesystem")
	unsetFlag := flag.Bool("unset", false, "Unset the root filesystem")
	infoFlag := flag.Bool("info", false, "Show information about the root filesystem")
	modeFlag := flag.String("mode", "", "Run command in a specific mode")
	toggleEmbeddedFlag := flag.Bool("toggle-embedded-bwrap", false, "Toggle between embedded and system bwrap")
	setModeFlagsFlag := flag.String("set-mode-flags", "", "Set flags for a specific mode (e.g., --set-mode-flags mode:\"flags\")")
	removeModeFlag := flag.String("remove-mode", "", "Remove a specific mode")
	setSedimentFlag := flag.String("sediment", "", "Set a mode to read-only (sediment)")
	help := &ccmd.CmdInfo{
		Name:        "noroot-do",
		Authors:     []string{"xplshn"},
		Repository:  "https://github.com/xplshn/a-utils",
		Description: "Interacts with a rootfs or container via leveraging bwrap",
		Synopsis:    "<|--set [ROOTFS_DIR]|--unset|--info|--mode [MODE] [CMD]|--toggle-embedded-bwrap|--set-mode-flags [name:\"--flag1 ...\"|--remove-mode [MODE]]|>",
		ExcludeFlags: map[string]bool{
			"sediment": true,
		},
		//CustomFields: map[string]interface{}{
		//	"1_Behavior": "Behavior is not set",
		//},
	}

	helpPage, err := help.GenerateHelpPage()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error generating help page:", err)
		os.Exit(1)
	}
	flag.Usage = func() {
		fmt.Print(helpPage)
	}

	flag.Parse()
	args := flag.Args()

	if *setFlag != "" {
		cmdSet(*setFlag)
	} else if *unsetFlag {
		cmdUnset()
	} else if *infoFlag {
		cmdInfo()
	} else if *toggleEmbeddedFlag {
		cmdToggleEmbeddedBwrap()
	} else if *modeFlag != "" {
		cmdRun(*modeFlag, args)
	} else if *setModeFlagsFlag != "" {
		parts := strings.SplitN(*setModeFlagsFlag, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid format for --set-mode-flags. Use --set-mode-flags modeName:\"flags to be passed to bwrap\"")
		}
		cmdSetModeFlags(parts[0], parts[1])
	} else if *removeModeFlag != "" {
		cmdRemoveMode(*removeModeFlag)
	} else if *setSedimentFlag != "" {
		cmdSetSediment(*setSedimentFlag)
	} else {
		flag.Usage()
        return
	}
}
