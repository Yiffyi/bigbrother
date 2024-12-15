#!/bin/bash

CGO_CFLAGS="-O2" go build -v -trimpath -ldflags="-s -w" --buildmode=c-shared -o bin/pam_bb.so ./exe/pam_bb
go build -v -trimpath -ldflags="-s -w" -o bin/bb ./exe/bb