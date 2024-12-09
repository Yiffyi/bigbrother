#!/bin/bash

CGO_CFLAGS="-g -O2" go build --buildmode=c-shared -o pam_bb.so ./pam_bb