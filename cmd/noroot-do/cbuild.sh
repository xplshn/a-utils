#!/bin/sh

OPWD="$PWD"
BASE="$(dirname "$(realpath "$0")")"
if [ "$OPWD" != "$BASE" ]; then
    echo "... $BASE is not the same as $PWD ..."
    echo "Going into $BASE and coming back here in a bit"
    cd "$BASE" || exit 1
fi
trap 'cd "$OPWD"' EXIT

# Function to log to stdout with green color
log() {
    _reset="\033[m"
    _blue="\033[34m"
    printf "${_blue}->${_reset} %s\n" "$*"
}

# Function to log_warning to stdout with yellow color
log_warning() {
    _reset="\033[m"
    _yellow="\033[33m"
    printf "${_yellow}->${_reset} %s\n" "$*"
}

# Function to log_error to stdout with red color
log_error() {
    _reset="\033[m"
    _red="\033[31m"
    printf "${_red}->${_reset} %s\n" "$*"
    exit 1
}

unnappear() {
    "$@" >/dev/null 2>&1
}

# Check if a dependency is available.
available() {
    unnappear which "$1" || return 1
}

# Exit if a dependency is not available
require() {
    available "$1" || log_error "[$1] is not installed. Please ensure the command is available [$1] and try again."
}

download_bwrap() {
    log "Downloading bwrap"
    if ! wget -qO "./bwrap" "https://bin.ajam.dev/$(uname -m)/bwrap"; then
        log_error "Unable to download bwrap"
    fi
    chmod +x "./bwrap"
}

build_project() {
    ROOT="--uid 0 --gid 0"
    SANDBOX="--unshare-all --share-net --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf"
    RSANDBOX="--share-net --proc /proc --dev-bind /dev /dev --bind /run /run --bind /tmp /tmp --ro-bind-try /usr/share/fonts /usr/share/fonts --ro-bind-try /usr/share/themes /usr/share/themes --ro-bind-try /sys /sys --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/machine-id /etc/machine-id --ro-bind-try /etc/asound.conf /etc/asound.conf --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /usr/share/fontconfig /usr/share/fontconfig"
    DESKTOP="--share-net --proc /proc --dev-bind /dev /dev --bind /run /run --bind /home /home --bind /tmp /tmp --bind-try /media /media --bind-try /mnt /mnt --bind-try /opt /opt --ro-bind-try /usr/share/fonts /usr/share/fonts --ro-bind-try /usr/share/themes /usr/share/themes --ro-bind-try /sys /sys --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/machine-id /etc/machine-id --ro-bind-try /etc/asound.conf /etc/asound.conf --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /usr/share/fontconfig /usr/share/fontconfig"
    go build || log_error "Go build failed"
    # SANDBOX
    log 'Creating "sandbox" preset'
    unnappear ./noroot-do --set-mode-flags sandbox:"$SANDBOX"
    unnappear ./noroot-do --sediment sandbox && log 'Sedimented "sandbox" preset'
    log 'Creating "rootsandbox" preset'
    unnappear ./noroot-do --set-mode-flags rootsandbox:"$SANDBOX $ROOT"
    unnappear ./noroot-do --sediment rootsandbox && log 'Sedimented "rootsandbox" preset'
    # RELAXED SANDBOX
    log 'Creating "relaxedSandbox" preset'
    unnappear ./noroot-do --set-mode-flags sandbox:"$RSANDBOX"
    unnappear ./noroot-do --sediment sandbox && log 'Sedimented "relaxedSandbox" preset'
    log 'Creating "rootRelaxedSandbox" preset'
    unnappear ./noroot-do --set-mode-flags rootsandbox:"$RSANDBOX $ROOT"
    unnappear ./noroot-do --sediment rootsandbox && log 'Sedimented "rootRelaxedSandbox" preset'
    # DESKTOP
    log 'Creating "desktop" preset'
    unnappear ./noroot-do --set-mode-flags desktop:"$DESKTOP"
    unnappear ./noroot-do --sediment desktop && log 'Sedimented "desktop" preset'
    log 'Creating "rootdesktop" preset'
    unnappear ./noroot-do --set-mode-flags rootdesktop:"$DESKTOP $ROOT"
    unnappear ./noroot-do --sediment rootdesktop && log 'Sedimented "rootdesktop" preset'
}

clean_project() {
    log "Starting clean process"
    unnappear rm ./noroot-do
    echo "rm ./noroot-do"
    unnappear rm ./bwrap
    echo "rm ./bwrap"
    log "Clean process completed"
}

retrieve_executable() {
    readlink -f ./noroot-do
}

# Main case statement for actions
case "$1" in
    "" | "build")
        require go
        log "Starting build process"
        [ -f "./bwrap" ] || download_bwrap
        build_project
        ;;
    "clean")
        clean_project
        ;;
    "retrieve")
        retrieve_executable
        ;;
    *)
        log_warning "Usage: $0 {build|clean|retrieve}"
        exit 1
        ;;
esac
