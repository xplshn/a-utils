#!/bin/sh

POSIXLY_CORRECT=1
build_OPWD="$PWD"
BASE="$(realpath "$0")" && BASE="${BASE%/*}"
if [ "$OPWD" != "$BASE" ]; then
    cd "$BASE" || log_error "Unable to change directory to ${BASE##*/}. Re-execute using a POSIX shell and check again."
fi
trap 'cd "$build_OPWD"' EXIT

# CONFIGURABLE PART
LOCAL_BUILD_DIR="$BASE/cmd"
LOCAL_BUILT_DIR="$BASE/built/bin"
EXCLUDE_BINARIES="demo getconf"  # Space-separated list of binaries to avoid building

# Function to log to stdout with green color
log() {
    _Xashstd_reset="\033[m"
    _Xashstd_color_code="\033[32m"
    printf "${_Xashstd_color_code}->${_Xashstd_reset} %s\n" "$*"
}

# Function to log_warning to stdout with yellow color
log_warning() {
    _Xashstd_reset="\033[m"
    _Xashstd_color_code="\033[33m"
    printf "${_Xashstd_color_code}->${_Xashstd_reset} %s\n" "$*"
}

# Function to log_error to stdout with red color
log_error() {
    _Xashstd_reset="\033[m"
    _Xashstd_color_code="\033[31m"
    printf "${_Xashstd_color_code}->${_Xashstd_reset} %s\n" "$*"
    exit 1
}

# Function to build Go commands in the specified directories
build_commands() {
    for main_dir in "$@"; do
        if [ -d "$main_dir" ]; then
            log "Processing directory: $main_dir"
            # Process directories containing .go files
            find "$main_dir" -type d | while IFS= read -r dir; do
                if [ -d "$dir" ]; then
                    # Check for .go files in the current directory
                    if find "$dir" -maxdepth 1 -type f -name '*.go' -print -quit | grep -q .; then
                        binary_name=$(basename "$dir")
                        if ! echo "$EXCLUDE_BINARIES" | grep -qw "$binary_name"; then
                            log "Building Go project in \"$dir\""
                            # Build the binary to a temporary directory
                            temp_bin_dir=$(mktemp -d)
                            go build -o "$temp_bin_dir/$binary_name" "$dir"
                            # Move the binary to the target directory
                            mv "$temp_bin_dir/$binary_name" "$LOCAL_BUILT_DIR"
                            rm -rf "$temp_bin_dir"
                        else
                            log_warning "Skipping excluded binary: $binary_name"
                        fi
                    fi
                elif [ "$main_dir" != "" ]; then
                    log_warning "Directory \"$dir\" does not exist"
                fi
            done
        elif [ "$main_dir" != "" ]; then
            log_warning "Directory \"$main_dir\" does not exist"
        fi
    done
}

# Function to clean up the built directory and remove binaries in specified main directories
clean_up() {
    # Remove the built directory
    log "Remove $(dirname "$LOCAL_BUILT_DIR")"
    rm -rf "$(dirname "$LOCAL_BUILT_DIR")" >/dev/null 2>&1
    # Process each main directory
    for main_dir in "$@"; do
        if [ -d "$main_dir" ]; then
            log "Processing directory: $main_dir"
            # Find all directories
            find "$main_dir" -type d | while IFS= read -r dir; do
                # Check for Go projects and clean them
                if [ -d "$dir" ] && find "$dir" -maxdepth 1 -type f -name '*.go' -print -quit | grep -q .; then
                    log "Cleaning Go project in $dir"
                    (cd "$dir" && go clean)
                fi
            done
        elif [ "$main_dir" != "" ]; then
            log_warning "Directory \"$main_dir\" does not exist"
        fi
    done
}

# Function to update dependencies for Go projects
update_go_deps() {
    for main_dir in "$@"; do
        if [ -d "$main_dir" ]; then
            log "Searching for Go projects in: $main_dir"
            # Find directories containing .go files
            find "$main_dir" -type d | while IFS= read -r dir; do
                if [ -d "$dir" ] && find "$dir" -maxdepth 1 -type f -name '*.go' -print -quit | grep -q .; then
                    log "Updating Go deps for: $dir"
                    (cd "$dir" && go get -u && go mod tidy)
                fi
            done
        elif [ "$main_dir" != "" ]; then
            log_warning "Directory \"$main_dir\" does not exist"
        fi
    done
}

case "$1" in
"" | "build")
    log "Starting build process"
    mkdir -p "$LOCAL_BUILT_DIR"
    build_commands "$LOCAL_BUILD_DIR"
    log "Build process completed"
    ;;
"clean")
    shift
    log "Starting clean process"
    clean_up "$LOCAL_BUILD_DIR"
    log "Clean process completed"
    ;;
"deps")
    log "Updating dependencies"
    update_go_deps "$LOCAL_BUILD_DIR"
    log "Dependencies updated"
    ;;
*)
    echo "Usage: $0 {build|clean|deps}"
    exit 1
    ;;
esac

# tar -czvf "./a-utils-$GOARCH-$GOOS.tar.gz" -C ./built/bin .
