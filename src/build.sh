#!/usr/bin/env bash
set -euo pipefail

cd -- "$(dirname -- "${BASH_SOURCE[0]}")"
mkdir -p dist

# Target the CC2's ARMv7 CPU without C libraries or debug metadata.
# Keep Go's default optimizations; avoid executable packers and their runtime cost.
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build \
    -trimpath \
    -buildvcs=false \
    -ldflags='-s -w -buildid=' \
    -o dist/elegoo-cc2-helper \
    .

printf 'Built dist/elegoo-cc2-helper (%s bytes)\n' "$(wc -c < dist/elegoo-cc2-helper)"
