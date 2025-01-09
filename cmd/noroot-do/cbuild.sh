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
    if ! wget -qO "./bwrap" "https://bin.pkgforge.dev/$(uname -m)/bwrap-patched"; then
        log_error "Unable to download bwrap"
    fi
    chmod +x "./bwrap"
}

build_project() {
    ROOT="--uid 0 --gid 0"
    SANDBOX="--unshare-all --proc /proc --share-net --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf --bind-try /home /home"
    RSANDBOX="--share-net --proc /proc --dev-bind /dev /dev --bind /run /run --bind /tmp /tmp --ro-bind-try /usr/share/fonts /usr/share/fonts --ro-bind-try /usr/share/themes /usr/share/themes --ro-bind-try /sys /sys --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/machine-id /etc/machine-id --ro-bind-try /etc/asound.conf /etc/asound.conf --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /usr/share/fontconfig /usr/share/fontconfig --bind-try /home /home"
    DESKTOP="--dev-bind-try / /_ --share-net --proc /proc --dev-bind /dev /dev --bind /run /run --bind /home /home --bind /tmp /tmp --bind-try /media /media --bind-try /mnt /mnt --bind-try /opt /opt --ro-bind-try /usr/share/fonts /usr/share/fonts --ro-bind-try /usr/share/themes /usr/share/themes --ro-bind-try /sys /sys --ro-bind-try /etc/resolv.conf /etc/resolv.conf --ro-bind-try /etc/hosts /etc/hosts --ro-bind-try /etc/nsswitch.conf /etc/nsswitch.conf --ro-bind-try /etc/passwd /etc/passwd --ro-bind-try /etc/group /etc/group --ro-bind-try /etc/machine-id /etc/machine-id --ro-bind-try /etc/asound.conf /etc/asound.conf --ro-bind-try /etc/localtime /etc/localtime --ro-bind-try /etc/hostname /etc/hostname --ro-bind-try /usr/share/fontconfig /usr/share/fontconfig"
    go build || log_error "Go build failed"
    # SANDBOX
    log 'Creating "sandbox" preset'
    unnappear ./noroot-do --set-mode-flags sandbox:"$SANDBOX"
    unnappear ./noroot-do --sediment sandbox && log 'Sedimented "sandbox" preset'
    log 'Creating "rootSandbox" preset'
    unnappear ./noroot-do --set-mode-flags rootSandbox:"$SANDBOX $ROOT"
    unnappear ./noroot-do --sediment rootSandbox && log 'Sedimented "rootSandbox" preset'
    # RELAXED SANDBOX
    log 'Creating "relaxedSandbox" preset'
    unnappear ./noroot-do --set-mode-flags relaxedSandbox:"$RSANDBOX"
    unnappear ./noroot-do --sediment relaxedSandbox && log 'Sedimented "relaxedSandbox" preset'
    log 'Creating "rootRelaxedSandbox" preset'
    unnappear ./noroot-do --set-mode-flags rootRelaxedSandbox:"$RSANDBOX $ROOT"
    unnappear ./noroot-do --sediment rootRelaxedSandbox && log 'Sedimented "rootRelaxedSandbox" preset'
    # DESKTOP
    log 'Creating "desktop" preset'
    unnappear ./noroot-do --set-mode-flags desktop:"$DESKTOP"
    unnappear ./noroot-do --sediment desktop && log 'Sedimented "desktop" preset'
    log 'Creating "rootDesktop" preset'
    unnappear ./noroot-do --set-mode-flags rootDesktop:"$DESKTOP $ROOT"
    unnappear ./noroot-do --sediment rootDesktop && log 'Sedimented "rootDesktop" preset'
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

# --dev-bind-try /System/Devices                 /dev
# #--dev-bind-try /System/Configuration           /etc
# --dev-bind-try /Users                          /home
# --dev-bind-try /System/Filesystems/External    /media
# --dev-bind-try /System/Filesystems/Internal    /mnt
# --dev-bind-try /System/Binaries/Optional       /opt
# #--dev-bind-try /System/Binaries/System         /bin
# #--dev-bind-try /System/Binaries/Standard       /usr/bind 
# #--dev-bind-try /System/Binaries/Administrative /usr/sbin 
# #--dev-bind-try /System/Libraries/System        /lib
# #--dev-bind-try /System/Libraries/Standard      /usr/lib
# #--dev-bind-try /System/Shareable               /usr/share
# --dev-bind-try /System/Variable                /var
