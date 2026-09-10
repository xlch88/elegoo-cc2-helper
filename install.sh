#!/bin/sh
set -eu

if [ "$#" -ne 1 ] || [ -z "$1" ]; then
    echo "Usage: sh install.sh PRINTER_IP (run on your Linux/macOS computer)" >&2
    exit 1
fi

command -v ssh >/dev/null 2>&1 || { echo "ssh is required" >&2; exit 1; }
command -v scp >/dev/null 2>&1 || { echo "scp is required" >&2; exit 1; }

URL="https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
BIN=$(mktemp "${TMPDIR:-/tmp}/elegoo-cc2-helper.XXXXXX")
trap 'rm -f "$BIN"' 0
trap 'exit 1' HUP INT TERM
REMOTE="/tmp/${BIN##*/}"

if command -v curl >/dev/null 2>&1; then
    curl -fL -o "$BIN" "$URL"
elif command -v wget >/dev/null 2>&1; then
    wget -O "$BIN" "$URL"
else
    echo "curl or wget with HTTPS support is required" >&2
    exit 1
fi

[ -s "$BIN" ] || { echo "Downloaded file is empty" >&2; exit 1; }
scp "$BIN" "root@$1:$REMOTE"
ssh "root@$1" "trap 'rm -f $REMOTE' 0; chmod +x $REMOTE && $REMOTE --install"
