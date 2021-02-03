#! /bin/bash
# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

# see: https://github.com/go-delve/delve/issues/865#issuecomment-480766102
if [ "$1" == "-d" ]; then
    go build -gcflags='all=-N -l' -o token-dice
else
    go build -o token-dice
fi

mkdir -p $SCRIPT_DIR/../bin/
mv token-dice $SCRIPT_DIR/../bin/
