package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/maja42/ember"
	"github.com/maja42/ember/embedding"
)

//go:embed bwrap
var embeddedBwrap embed.FS

// Configuration stores all the program settings
type Configuration struct {
	RootFS        string            `json:"rootfs"`
	UseEmbedded   bool              `json:"use_embedded_bwrap"`
	ModeFlags     map[string]string `json:"mode_flags"`
	SedimentModes map[string]bool   `json:"sediment_modes"`
}

// Global configuration
var config Configuration

// Initialize default configuration
func initDefaultConfig() Configuration {
	return Configuration{
		RootFS:        "",
		UseEmbedded:   false,
		ModeFlags:     make(map[string]string),
		SedimentModes: make(map[string]bool),
	}
}

// Load configuration from embedded data
func loadConfig() error {
	attachments, err := ember.Open()
	if err == nil {
		defer attachments.Close()

		contents := attachments.List()
		hasConfig := false
		for _, name := range contents {
			if name == "config.json" {
				hasConfig = true
				break
			}
		}

		if hasConfig {
			r := attachments.Reader("config.json")
			data, err := io.ReadAll(r)
			if err != nil {
				return fmt.Errorf("failed to read config: %v", err)
			}
			if err := json.Unmarshal(data, &config); err != nil {
				return fmt.Errorf("failed to parse config: %v", err)
			}
			return nil
		}
	}
	// If no config exists, initialize defaults
	config = initDefaultConfig()
	return nil
}

func saveConfig() error {
	// Marshal the current config
	data, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}

	// Get the path to the current executable
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}

	// Create a backup of the current executable first
	backupPath := executable + ".backup"
	if err := copyFile(executable, backupPath); err != nil {
		return fmt.Errorf("failed to create backup: %v", err)
	}

	// Create all temporary files in the same directory as the executable
	// This ensures atomic rename operations work across filesystems
	execDir := filepath.Dir(executable)

	// Create a temporary file for the config
	configTempFile, err := os.CreateTemp(execDir, "config-*.json")
	if err != nil {
		return fmt.Errorf("failed to create temp config file: %v", err)
	}
	defer os.Remove(configTempFile.Name())

	// Write the config data to the temporary file
	if err := os.WriteFile(configTempFile.Name(), data, 0644); err != nil {
		return fmt.Errorf("failed to write temp config: %v", err)
	}

	// Create a temporary file for the new executable
	newExePath := executable + ".new"
	newExe, err := os.OpenFile(newExePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to create new executable file: %v", err)
	}
	defer newExe.Close()
	defer os.Remove(newExePath) // Clean up in case of failure

	// Open the backup file as source
	sourceExe, err := os.Open(backupPath)
	if err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to open backup: %v", err)
	}
	defer sourceExe.Close()

	// Setup logging function
	logger := func(format string, args ...interface{}) {
		fmt.Printf("\t"+format+"\n", args...)
	}

	// Create a temporary file for the cleaned executable
	cleanedExePath := executable + ".cleaned"
	cleanedExe, err := os.OpenFile(cleanedExePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to create cleaned executable file: %v", err)
	}
	defer cleanedExe.Close()
	defer os.Remove(cleanedExePath)

	// Try to remove existing embedding, but don't fail if there isn't any
	err = embedding.RemoveEmbedding(cleanedExe, sourceExe, logger)
	if err != nil && !strings.Contains(err.Error(), "contains no embedded data") {
		os.Remove(backupPath)
		return fmt.Errorf("failed to remove existing embedding: %v", err)
	}

	// Determine which file to use as source
	var embedSource string
	if err == nil {
		embedSource = cleanedExePath
	} else {
		embedSource = backupPath
	}

	// Open the source file for embedding
	embedSourceFile, err := os.Open(embedSource)
	if err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to open source for embedding: %v", err)
	}
	defer embedSourceFile.Close()

	// Define attachments for the new config
	attachments := map[string]string{
		"config.json": configTempFile.Name(),
	}

	// Create the new executable with the embedded config
	if err := embedding.EmbedFiles(newExe, embedSourceFile, attachments, logger); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to embed new config: %v", err)
	}

	// Close all files before attempting rename
	newExe.Close()
	embedSourceFile.Close()
	cleanedExe.Close()

	// Atomic replacement of the executable
	if err := os.Rename(newExePath, executable); err != nil {
		// If rename fails, restore from backup
		if restoreErr := os.Rename(backupPath, executable); restoreErr != nil {
			return fmt.Errorf("failed to update executable and restore backup: %v (original error: %v)", restoreErr, err)
		}
		return fmt.Errorf("failed to update executable (restored from backup): %v", err)
	}

	// Remove the backup file only after successful update
	os.Remove(backupPath)

	return nil
}

// Run a command inside the chroot environment using bwrap
func runBwrapCommand(args []string, mode string) error {
	if len(args) == 0 {
		return fmt.Errorf("no command provided to run inside the chroot")
	}

	bwrapCmd := "bwrap"

	if config.UseEmbedded {
		tmpFile, err := os.CreateTemp("", "bwrap-embedded-*")
		if err != nil {
			return fmt.Errorf("failed to create temp file for embedded bwrap: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		bwrapBinary, err := embeddedBwrap.Open("bwrap")
		if err != nil {
			return fmt.Errorf("failed to open embedded bwrap binary: %v", err)
		}
		defer bwrapBinary.Close()

		if _, err := io.Copy(tmpFile, bwrapBinary); err != nil {
			return fmt.Errorf("failed to write embedded bwrap to temp file: %v", err)
		}

		if err := tmpFile.Close(); err != nil {
			return fmt.Errorf("failed to close temp file: %v", err)
		}

		if err := os.Chmod(tmpFile.Name(), 0755); err != nil {
			return fmt.Errorf("failed to make temp bwrap executable: %v", err)
		}

		bwrapCmd = tmpFile.Name()
	}

	bwrapArgs := []string{
		"--bind", config.RootFS, "/",
	}

	// Add mode-specific flags
	if flags, ok := config.ModeFlags[mode]; ok {
		bwrapArgs = append(bwrapArgs, strings.Split(flags, " ")...)
	}

	bwrapArgs = append(bwrapArgs, "--")
	bwrapArgs = append(bwrapArgs, args...)

	cmd := exec.Command(bwrapCmd, bwrapArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// Command handlers
func cmdSet(rootfs string) {
	absRootfs, err := filepath.Abs(rootfs)
	if err != nil {
		log.Fatalf("Error determining absolute path: %v", err)
	}
	config.RootFS = absRootfs
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Printf("Root filesystem set to: %s\n", absRootfs)
}

func cmdUnset() {
	if config.RootFS == "" {
		log.Fatalf("No rootfs is currently set.")
	}
	config.RootFS = ""
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Println("Root filesystem unset successfully.")
}

func cmdInfo() {
	fmt.Printf("rootfs: %s\n", config.RootFS)
	if config.UseEmbedded {
		fmt.Println("bwrap: embedded")
	} else {
		fmt.Println("bwrap: system")
	}

	for mode, flags := range config.ModeFlags {
		fmt.Printf("Mode: \x1b[94m%s\x1b[0m, Flags: %s\n", mode, flags)
		if config.SedimentModes[mode] {
			fmt.Printf("Mode %s is read-only (sedimented).\n", mode)
		}
	}
}

func cmdToggleEmbeddedBwrap() {
	config.UseEmbedded = !config.UseEmbedded
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	if config.UseEmbedded {
		fmt.Println("Now using embedded bwrap.")
	} else {
		fmt.Println("Now using system bwrap.")
	}
}

func cmdRun(mode string, args []string) {
	if config.RootFS == "" {
		log.Fatalf("No rootfs is currently set.")
	}
	if err := runBwrapCommand(args, mode); err != nil {
		log.Fatalf("Error running command in chroot: %v", err)
	}
}

func cmdSetModeFlags(mode string, flags string) {
	if config.SedimentModes[mode] {
		log.Fatalf("Mode %s is read-only (sedimented) and cannot be modified.", mode)
	}
	config.ModeFlags[mode] = flags
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Printf("Successfully configured mode \"%s\"\n", mode)
}

func cmdRemoveMode(mode string) {
	if config.SedimentModes[mode] {
		log.Fatalf("Mode %s is read-only (sedimented) and cannot be removed.", mode)
	}
	delete(config.ModeFlags, mode)
	delete(config.SedimentModes, mode)
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Printf("Mode %s removed successfully.\n", mode)
}

func cmdSetSediment(mode string) {
	if _, exists := config.ModeFlags[mode]; !exists {
		log.Fatalf("Mode %s does not exist.", mode)
	}
	config.SedimentModes[mode] = true
	if err := saveConfig(); err != nil {
		log.Fatalf("Error saving config: %v", err)
	}
	fmt.Printf("Mode %s set to read-only (sedimented).\n", mode)
}

func main() {
	if err := loadConfig(); err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	setFlag := flag.String("set", "", "Set the root filesystem")
	unsetFlag := flag.Bool("unset", false, "Unset the root filesystem")
	infoFlag := flag.Bool("info", false, "Show information about the root filesystem")
	modeFlag := flag.String("mode", "", "Run command in a specific mode")
	toggleEmbeddedFlag := flag.Bool("toggle-embedded-bwrap", false, "Toggle between embedded and system bwrap")
	setModeFlagsFlag := flag.String("set-mode-flags", "", "Set flags for a specific mode (e.g., --set-mode-flags mode:\"flags\")")
	removeModeFlag := flag.String("remove-mode", "", "Remove a specific mode")
	setSedimentFlag := flag.String("sediment", "", "Set a mode to read-only (sediment)")

	flag.Parse()
	args := flag.Args()

	switch {
	case *setFlag != "":
		cmdSet(*setFlag)
	case *unsetFlag:
		cmdUnset()
	case *infoFlag:
		cmdInfo()
	case *toggleEmbeddedFlag:
		cmdToggleEmbeddedBwrap()
	case *modeFlag != "":
		cmdRun(*modeFlag, args)
	case *setModeFlagsFlag != "":
		parts := strings.SplitN(*setModeFlagsFlag, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("Invalid format for --set-mode-flags. Use --set-mode-flags modeName:\"flags to be passed to bwrap\"")
		}
		cmdSetModeFlags(parts[0], parts[1])
	case *removeModeFlag != "":
		cmdRemoveMode(*removeModeFlag)
	case *setSedimentFlag != "":
		cmdSetSediment(*setSedimentFlag)
	default:
		flag.Usage()
	}
}

// copyFile copies a file from src to dst. If dst already exists, it will be overwritten.
func copyFile(src, dst string) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %v", err)
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %v", err)
	}
	defer destFile.Close()

	// Copy contents from source to destination
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy file contents: %v", err)
	}

	// Set permissions on the destination file (same as source)
	sourceFileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to get source file info: %v", err)
	}
	err = os.Chmod(dst, sourceFileInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to set destination file permissions: %v", err)
	}

	// Return success
	return nil
}
