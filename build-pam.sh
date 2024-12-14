#!/bin/bash

CGO_CFLAGS="-O2" go build -v -trimpath -ldflags="-s -w" --buildmode=c-shared -o pam_bb.so ./pam_bb