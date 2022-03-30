#!/bin/bash
# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
if [[ "" == "$1" ]]; then
    echo "Usage: gen-key.sh OUTFILE"
    exit 1
fi
set -x -e
openssl genpkey -algorithm EC -out $1 -pkeyopt ec_paramgen_curve:P-256
openssl ec -in $1 -pubout -out $1.pub
