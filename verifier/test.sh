#! /bin/bash
# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

pushd $SCRIPT_DIR/../plugins

if [ "$1" == "-d" ]; then
    ./buildall.sh -d
    popd
    dlv test
else
    ./buildall.sh
    popd
    go test
fi
