#!/bin/sh

POSIXLY_CORRECT=1
build_OPWD="$PWD"
BASE="$(realpath "$0")" && BASE="${BASE%/*}"
if [ "$OPWD" != "$BASE" ]; then
    cd "$BASE" || log "$R" "Unable to change directory to ${BASE##*/}. Re-execute using a POSIX shell and check again."
fi
trap 'cd "$build_OPWD"' EXIT


# CONFIGURABLE PART
EXTEND_FILES="$BASE/extendGo.b $BASE/extendC.b"
EXTEND_EFILES="$BASE/extendExclude.eb"
EXTEND_DIR="$BASE/extend"
LOCAL_BUILD_DIR="$BASE/cmd"
EXTEND_BUILT_DIR="$BASE/built/usr/bin"
LOCAL_BUILT_DIR="$BASE/built/bin"

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

# Unified function to build Go commands or execute cbuild.sh in the specified directories
build_commands() {
    for main_dir in "$@"; do
        if [ -d "$main_dir" ]; then
            log "Processing directory: $main_dir"
            # Process directories containing .go files or cbuild.sh
            find "$main_dir" -type d | while IFS= read -r dir; do
                if [ -d "$dir" ]; then
                    # Check for .go files in the current directory
                    if find "$dir" -maxdepth 1 -type f -name '*.go' -print -quit | grep -q .; then
                        log "Building Go projects in \"$dir\" and placing in $LOCAL_BUILT_DIR"
                        (cd "$dir" && GOBIN="$LOCAL_BUILT_DIR" go install .)
                    fi
                    # Check for cbuild.sh in the current directory
                    if [ -f "$dir/cbuild.sh" ]; then
                        log "Building C project in $dir"
                        (cd "$dir" && ./cbuild.sh)
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

# Function to move built executables to the built directory
move_executables() {
    # Find and move files listed by cbuild.sh scripts
    find "$@" -type f -name 'cbuild.sh' -print | while IFS= read -r cbuild_file; do
        # Execute the cbuild.sh script and handle its output
        "$cbuild_file" retrieve | while IFS= read -r file; do
            # Ensure that only files (not directories) are copied
            if [ -f "$file" ]; then
                # Move the file directly to $EXTEND_BUILT_DIR
                mv "$file" "$EXTEND_BUILT_DIR" && log "Moved \"$file\" to \"$EXTEND_BUILT_DIR\""
            fi
        done
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
            # Find all directories a cbuild.sh script
            find "$main_dir" -type d | while IFS= read -r dir; do
                # Execute cbuild.sh clean if it exists
                if [ -f "$dir/cbuild.sh" ]; then
                    log "Cleaning C project in $dir"
                    (cd "$dir" && ./cbuild.sh clean)
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

# Unified function to read extend files, clone or update repositories
process_repos() {
    target_dir=$1
    shift
    mkdir -p "$target_dir"
    # Process each file passed as arguments
    # shellcheck disable=SC2068 # We want to re-split the array expansion
    for extend_file in $@; do
        if [ -f "$extend_file" ]; then
            while IFS= read -r line; do
                # Remove leading and trailing whitespace
                line=$(echo "$line" | xargs)
                # Skip empty lines and lines starting with #
                if [ -z "$line" ] || echo "$line" | grep -q '^#'; then
                    continue
                fi

                # Extract REPO_DEST
                REPO_DEST=$(echo "$line" | awk -F '->' '{print $1}' | xargs)
                # Extract REPO_URL
                REPO_URL=$(echo "$line" | awk -F '->' '{print $2}' | awk -F '-<' '{print $1}' | xargs)
                # Extract GIT_CLONE_ARGS
                GIT_CLONE_ARGS=$(echo "$line" | awk -F '-<' '{print $2}' | awk -F '-><-' '{print $1}' | xargs)
                # Extract SH_CMD
                SH_CMD=$(echo "$line" | awk -F '-><-' '{print $2}' | xargs)

                # Process the repository destination
                if [ -z "$REPO_DEST" ]; then
                    echo "error: REPO_DEST cannot be empty!" >&2
                    continue
                fi
                REPO_DEST="$target_dir/$REPO_DEST"

                # Clone or update repository
                if [ -d "$REPO_DEST/.git" ]; then
                    echo "Updating repository in $REPO_DEST"
                    (cd "$REPO_DEST" && git pull)
                else
                    echo "Cloning repository into \"$REPO_DEST\""
                    if [ -n "$GIT_CLONE_ARGS" ]; then
                        log_warning "Passing \"$GIT_CLONE_ARGS\" to \"git clone\""
                        # shellcheck disable=SC2086 # We want word splitting
                        git clone --depth 1 $GIT_CLONE_ARGS "$REPO_URL" "$REPO_DEST"
                    else
                        git clone --depth 1 "$REPO_URL" "$REPO_DEST"
                    fi
                fi
                # Execute shell command if provided
                if [ -n "$SH_CMD" ]; then
                    (cd "$REPO_DEST" && $SH_CMD) || echo "Unable to execute \"$SH_CMD\" in $REPO_DEST" >&2
                fi
            done < "$extend_file"
        else
            echo "File not found: $extend_file" >&2
        fi
    done
}

# Function to remove excluded files listed in $1 from the "built" directory inside each target_dir
remove_excluded_files() {
    list_file="$1"
    shift

    if [ ! -f "$list_file" ]; then
        log_error "Error: $list_file does not exist."
    fi

    # Read the exclude list, ignoring comments and empty lines
    exclude_files=$(grep -v '^\s*#' "$list_file" | grep -v '^\s*$')

    # Iterate over each target directory passed as an argument
    # shellcheck disable=SC2068 # We want to re-split the elements
    for target_dir in $@; do
        # Iterate over each file in the exclude list
        for exclude_file in $exclude_files; do
            # Find the matching executable file in the "built" directory
            find "$target_dir" -type f -name "$exclude_file" | while IFS= read -r file; do
                rm -f "$file" || log_error "Failed to remove \"$exclude_file\" from \"$target_dir\"" && log_warning "Removed \"$exclude_file\" from \"$target_dir\""
            done
        done
    done
}

case "$1" in
"" | "build")
    log "Starting build process"
    process_repos "$EXTEND_DIR" "$EXTEND_FILES"
    mkdir -p "$LOCAL_BUILT_DIR" "$EXTEND_BUILT_DIR"
    build_commands "$LOCAL_BUILD_DIR" "$EXTEND_DIR"
    move_executables "$LOCAL_BUILD_DIR" "$EXTEND_DIR"
    remove_excluded_files "$EXTEND_EFILES" "$EXTEND_BUILT_DIR" "$LOCAL_BUILT_DIR"
    log "Build process completed"
    ;;
"clean")
    shift
    log "Starting clean process"
    clean_up "$EXTEND_DIR" # There is no procedure to clean Go directories in clean_up, that's why we don't pass $LOCAL_BUILD_DIR, which only contains Go files for now.
    if [ "$1" = "full" ]; then
        shift
        log "Removing $EXTEND_DIR directory"
        rm -rf ./extend >/dev/null 2>&1
    fi
    log "Clean process completed"
    ;;
"deps")
    log "Updating dependencies"
    update_go_deps "$LOCAL_BUILD_DIR" "$EXTEND_DIR"
    log "Dependencies updated"
    ;;
*)
    echo "Usage: $0 {build|clean <full>|deps}"
    exit 1
    ;;
esac
