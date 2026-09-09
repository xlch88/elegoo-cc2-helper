#!/bin/sh
set -eu

URL="https://github.com/xlch88/elegoo-cc2-helper/releases/latest/download/elegoo-cc2-helper"
BIN="/tmp/elegoo-cc2-helper"

if command -v uclient-fetch >/dev/null 2>&1; then
    uclient-fetch -O "$BIN" "$URL"
else
    echo "uclient-fetch is required for GitHub HTTPS downloads on this printer" >&2
    exit 1
fi

chmod +x "$BIN"
"$BIN" --install
