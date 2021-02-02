#!/bin/bash
# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0


if [ "$1" == "-d" ]; then
    OPTS="-d"
else
    OPTS=""
fi

for d in $(find . -maxdepth 1 -type d); do
	if [[ -f $d/build.sh ]]; then
		cd $d
                echo "Building $d..."
		. ./build.sh $OPTS
		cd ..
        fi
done
echo "Done."
