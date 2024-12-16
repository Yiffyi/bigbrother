#!/bin/bash

export CGO_CFLAGS="-Os"

go build -v -trimpath -ldflags="-s -w" --buildmode=c-shared -o bin/pam_bb.so ./exe/pam_bb
go build -v -trimpath -ldflags="-s -w" -o bin/honeypot ./exe/honeypot

# don't pack main executable
rm bin/bb exe/bb/install.tar exe/bb/install.tar.zst

tar -cf exe/bb/install.tar -C bin .
zstd -9 exe/bb/install.tar -o exe/bb/install.tar.zst

go build -v -trimpath -ldflags="-s -w" -o bin/bb ./exe/bb
