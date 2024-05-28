package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

/* <-BSD-3-Clause License->
Copyright <2024> <xplshn@murena.io>
Redistribution and use in source and binary forms, with or without modification, are permitted provided that the following conditions are met:
1. Redistributions of source code must retain the above copyright notice, this list of conditions and the following disclaimer.
2. Redistributions in binary form must reproduce the above copyright notice, this list of conditions and the following disclaimer in the documentation and/or other materials provided with the distribution.
3. Neither the name of the copyright holder nor the names of its contributors may be used to endorse or promote products derived from this software without specific prior written permission.
THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <ELF_SRC_PATH> <DST_PATH.blob> [--add-library <LIB_PATH>] [--add-binary <BIN_PATH>] [--add-arbitrary <DIR,FILE>]\n", os.Args[0])
		os.Exit(1)
	}

	src := os.Args[1]
	dst := os.Args[2]

	outerTmpDir, err := ioutil.TempDir("", "pelf_*")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create temporary directory: %v\n", err)
		os.Exit(1)
	}
	defer os.RemoveAll(outerTmpDir)

	tmpDir := filepath.Join(outerTmpDir, "pelf_tmp")
	err = os.MkdirAll(filepath.Join(tmpDir, "bin"), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directories: %v\n", err)
		os.Exit(1)
	}
	err = os.MkdirAll(filepath.Join(tmpDir, "libs"), 0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create directories: %v\n", err)
		os.Exit(1)
	}

	addTheLibs := func(binary string) error {
		cmd := exec.Command("ldd", binary)
		output, err := cmd.Output()
		if err != nil {
			return err
		}
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			parts := strings.Fields(line)
			if len(parts) >= 3 && parts[1] == "=>" && parts[2] != parts[0] {
				libPath := parts[2]
				err = copyFile(libPath, filepath.Join(tmpDir, "libs", filepath.Base(libPath)))
				if err != nil {
					return err
				}
			}
		}
		return nil
	}

	addLibrary := func(lib string) error {
		err := copyFile(lib, filepath.Join(tmpDir, "libs", filepath.Base(lib)))
		if err != nil {
			return err
		}
		return addTheLibs(lib)
	}

	addBinary := func(binary string) error {
		err := addTheLibs(binary)
		if err != nil {
			return err
		}
		return copyFile(binary, filepath.Join(tmpDir, "bin", filepath.Base(binary)))
	}

	addArbitrary := func(src string) error {
		return copyDir(src, tmpDir)
	}

	flagSet := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	var addLibraryPath, addBinaryPath, addArbitraryPath string
	flagSet.StringVar(&addLibraryPath, "add-library", "", "Path to library to add")
	flagSet.StringVar(&addBinaryPath, "add-binary", "", "Path to binary to add")
	flagSet.StringVar(&addArbitraryPath, "add-arbitrary", "", "Path to arbitrary file or directory to add")
	flagSet.Parse(os.Args[3:])

	if addLibraryPath != "" {
		err = addLibrary(addLibraryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add library: %v\n", err)
			os.Exit(1)
		}
	}

	if addBinaryPath != "" {
		err = addBinary(addBinaryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add binary: %v\n", err)
			os.Exit(1)
		}
	}

	if addArbitraryPath != "" {
		err = addArbitrary(addArbitraryPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to add arbitrary files: %v\n", err)
			os.Exit(1)
		}
	}

	err = addBinary(src)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to add binary: %v\n", err)
		os.Exit(1)
	}

	archivePath := filepath.Join(outerTmpDir, "archive.tar.gz")
	err = createTarGz(archivePath, tmpDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create archive: %v\n", err)
		os.Exit(1)
	}

	scriptFile, err := os.Create(dst)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create destination file: %v\n", err)
		os.Exit(1)
	}
	defer scriptFile.Close()

	scriptContent := `#!/bin/sh
# Get the binary's name
rEXE_NAME="$(basename "$0" .blob)"
[ -n "$EXE_NAME" ] || EXE_NAME="$rEXE_NAME"
TMPDIR="/tmp/.pelfbundles/pbundle_$rEXE_NAME$(date '+%s%M%S')_$RANDOM"
LIBS_BULKDIR="/tmp/pelfbundle_libs-bulkdir"
cleanup() {
    if [ -z "$found_runningInstance" ] || [ "$found_runningInstance" != "1" ]; then
        # Delete individual files only if they are used exclusively by the current process
        for file in $REM_AFTERUSE; do
            if [ -z "$(fuser "$file" 2>/dev/null | grep "$EXE_NAME_PID")" ]; then
                rm "$file"
            fi
        done

        # Delete the directory
        rm -rf "$TMPDIR"
    fi
}
# Set up the trap
trap cleanup EXIT
###########################################################

set_tmpdir_from_env() {
    # Construct the variable name by appending _bindir to $rEXENAME
    var_name="${rEXE_NAME}_bindir"

    # Check if the constructed variable name exists and is not empty
    eval "var_value=\${$var_name}"
    if [ -d "$var_value" ]; then
        # Set TMPDIR to the directory name of the constructed variable
        TMPDIR="$(dirname "$var_value")"
        found_runningInstance=1
        return
    fi
}

set_tmpdir_from_env
if [ -z "$found_runningInstance" ] || [ "$found_runningInstance" != "1" ]; then
        # Find the start position of the archive
        ARCHIVE_MARKER=$(awk '/^__ARCHIVE_MARKER__/ { print NR + 1; exit }' "$0")

        # Construct the variable name by appending _bindir to $rEXENAME
        var_name="${rEXE_NAME}_bindir"
        # Decode the base64-encoded archive and extract it
        mkdir -p "$TMPDIR" && tail -n +$ARCHIVE_MARKER "$0" | base64 -d | tar -xzf - -C "$TMPDIR" >/dev/null 2>&1 || {
            # Use eval to check if the constructed variable name exists and is not empty
            echo "Extraction failed" >&2
            eval "var_value=\"\${$var_name}\""
            exit 1
        }
fi

# Function to check if a library is found in system paths
is_library_in_system() {
    library=$1
    if [ -e "/usr/lib/$library" ] || [ -e "/lib/$library" ] || [ -e "/lib64/$library" ]; then
        return 0 # Library found in system
    else
        return 1 # Library not found in system
    fi
}

# Check if USE_SYSTEM_LIBRARIES is set to 1 or doesn't exist
if [ "${USE_SYSTEM_LIBRARIES:-0}" -eq 1 ]; then
    for lib_file in "$TMPDIR/libs/"*; do
        lib_name=$(basename "$lib_file")

        if is_library_in_system "$lib_name"; then
            if [ "$SHOW_DISCARDPROCESS" -eq 1 ]; then
                echo "$lib_name found in system. Using the system's library."
            fi
            rm "$lib_file"
        else
            if [ "$SHOW_DISCARDPROCESS" -eq 1 ]; then
                echo "$lib_name not found in system. Using the bundl
ed library."
            fi
        fi
    done 2>/dev/null
fi

mv_u() {
  SRC_DIR="$1"
  DEST_DIR="$2"

  # Loop through each file in the source directory
  for file in "$SRC_DIR"/*; do
    # Check if the file is a regular file
    [ -f "$file" ] || continue
    # Extract the filename from the path
    filename=$(basename "$file")
    # Check if the file does not exist in the destination directory or is newer
    if [ ! -e "$DEST_DIR/$file" ]; then
      REM_AFTERUSE="$REM_AFTERUSE $DEST_DIR/$filename "
      mv "$file" "$DEST_DIR/"
    elif [ "$(find "$file" -newer "$DEST_DIR/$filename" 2>/dev/null)" ]; then
      # Move the file to the destination directory
      mv "$file" "$DEST_DIR/"
    fi
  done
}

# Add extra binaries to the PATH, if they are there.
if [ "$(ls -1 "$TMPDIR"/bin | wc -l)" -gt 1 ]; then
        if [ -z "$found_runningInstance" ] || [ "$found_runningInstance" != "1" ]; then
                        export "$(echo "$rEXE_NAME" | sed -E 's/[-.]([a-zA-Z])/\U\1/g; s/[-.]//g')_bindir"="$TMPDIR/bin"
                        export "$(echo "$rEXE_NAME" | sed -E 's/[-.]([a-zA-Z])/\U\1/g; s/[-.]//g')_libs"="$TMPDIR/libs"
        fi
        xPATH="$TMPDIR/bin:$PATH"
        USE_BULKLIBS=0
fi

# Figure out what we do
binDest="$TMPDIR/bin/$EXE_NAME"
if [ "$1" = "--pbundle_help" ]; then
    printf "Description: Pack an ELF\n"
    printf "Usage:\n <--pbundle_link <binary>|--pbundle_help> <args...>\n"
    printf "EnvVars:\n USE_BULKLIBS=[0,1]\n USE_SYSTEM_LIBRARIES=[1,0]\n SHOW_DISCARDPROCESS=[0,1]\n HELP_PAGE_LIST_PACKEDFILES=[0,1]\n"
    if [ "$HELP_PAGE_LIST_PACKEDFILES" = "1" ]; then
        ls "$TMPDIR"/*
    fi
    exit 1
fi
if [ "$1" = "--pbundle_link" ]; then
    binDest="$2"
    shift 2
fi

# Execute the binary with extracted libraries using LD_LIBRARY_PATH
if [ "${USE_BULKLIBS:-0}" -eq 1 ]; then
   mkdir -p "$LIBS_BULKDIR"
   mv_u "$TMPDIR/libs" "$LIBS_BULKDIR"
   PATH="$PATH:$xPATH" LD_LIBRARY_PATH="$LIBS_BULKDIR" SELF_TEMPDIR="$TMPDIR" "$binDest" "$@" || exit 1
   EXE_NAME_PID="$!"
else
   PATH="$PATH:$xPATH" LD_LIBRARY_PATH="$TMPDIR/libs" SELF_TEMPDIR="$TMPDIR" "$binDest" "$@" || exit 1
   EXE_NAME_PID="$!"
fi

exit $?
__ARCHIVE_MARKER__
`
	_, err = scriptFile.WriteString(scriptContent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write script content: %v\n", err)
		os.Exit(1)
	}

	archiveFile, err := os.Open(archivePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open archive: %v\n", err)
		os.Exit(1)
	}
	defer archiveFile.Close()

	encoder := base64.NewEncoder(base64.StdEncoding, scriptFile)
	_, err = io.Copy(encoder, archiveFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to encode archive: %v\n", err)
		os.Exit(1)
	}
	encoder.Close()

	err = scriptFile.Chmod(0755)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to make script executable: %v\n", err)
		os.Exit(1)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	err = out.Close()
	if err != nil {
		return err
	}

	fi, err := in.Stat()
	if err != nil {
		return err
	}

	return os.Chmod(dst, fi.Mode())
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(path, src)
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, dstPath)
		}

		err = copyFile(path, dstPath)
		if err != nil {
			return err
		}

		return os.Chmod(dstPath, info.Mode())
	})
}

func createTarGz(dst, srcDir string) error {
	file, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	return filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath := strings.TrimPrefix(file, srcDir)
		if relPath == "" {
			return nil
		}

		header, err := tar.FileInfoHeader(fi, relPath)
		if err != nil {
			return err
		}

		header.Name = relPath
		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(tarWriter, f)
			if err != nil {
				return err
			}
		}

		return nil
	})
}
