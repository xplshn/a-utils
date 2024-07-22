#!/bin/sh

OPWD="$PWD"
BASE="$(dirname "$(realpath "$0")")"
if [ "$OPWD" != "$BASE" ]; then
    echo "... $BASE is not the same as $PWD ..."
    echo "Going into $BASE and coming back here in a bit"
    cd "$BASE" || exit 1
fi
trap 'cd "$OPWD"' EXIT

# Function to build each command in the cmd directory
build_commands() {
    for dir in ./cmd/*; do
        if [ -d "$dir" ]; then
            echo "Building $dir"
            (cd "$dir" && go build)
        fi
    done
}

# Function to copy built executables to the built directory
copy_executables() {
    mkdir -p ./built
    find ./cmd -type f -executable | while read -r executable; do
        mv "$executable" ./built/
    done
}

# Function to clean up the built directory and remove binaries in ./built
clean_up() {
    case "$1" in
        "full")
            rm -rf ./built
            find ./cmd -type f -executable -exec rm {} +
            ;;
        *)
            echo "Invalid argument. Use 'full' to clean everything."
            ;;
    esac
}

# Function to update dependencies
update_deps() {
    for dir in ./cmd/*; do
        if [ -d "$dir" ]; then
            echo "Updating deps for: $dir"
            (cd "$dir" && go get -u ; go mod tidy)
        fi
    done
}

# Main script execution
case "$1" in
        ""|"build")
        build_commands
        copy_executables
        ;;
    "clean")
        clean_up "$2"
        ;;
    "deps")
        update_deps
        ;;
        *)
        echo "Usage: $0 {--help} {build|clean|deps}"
        exit 1
        ;;
esac
