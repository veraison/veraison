#!/bin/bash
# Copyright 2020-2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

# Following command brings up the docker!
make up

loop_count=10
while [ "$(curl -s -L -o /dev/null -w "%{http_code}" http://localhost:8529)" != 200 -o $loop_count -lt 0 ] 
  do sleep 1 ;
  loop_count=$(( loop_count-1 ))
  echo "Waiting for Arango to come up"
done
echo "Hurray, arango is up"
