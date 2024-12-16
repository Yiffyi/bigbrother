#!/bin/bash
set -e
export CGO_CFLAGS="-Os"

echo "-- [ Building exe/pam_bb ] --"
go build -v -trimpath -ldflags="-s -w" --buildmode=c-shared -o bin/pam_bb.so ./exe/pam_bb

echo "-- [ Building exe/honeypot ] --"
go build -v -trimpath -ldflags="-s -w" -o bin/honeypot ./exe/honeypot

# don't pack main executable
# it's OK that those doesn't exists
rm bin/bb exe/bb/install.tar exe/bb/install.tar.zst || true

tar -cf exe/bb/install.tar -C bin .
zstd -9 exe/bb/install.tar -o exe/bb/install.tar.zst

echo "-- [ Building exe/bb ] --"
go build -v -trimpath -ldflags="-s -w" -o bin/bb ./exe/bb
