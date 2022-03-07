#!/bin/bash
# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

set -e

type go-licenses &> /dev/null || go get github.com/google/go-licenses

MODULES+=("cmd/vts github.com/veraison/cmd/trustedservices")
MODULES+=("cmd/endorsements github.com/veraison/cmd/endorsements")
MODULES+=("cmd/policy github.com/veraison/cmd/policy")
MODULES+=("endorsement github.com/veraison/endorsement")
MODULES+=("frontend github.com/veraison/frontend")
MODULES+=("plugins/arangodbendorsements arangodbendorsements")
MODULES+=("plugins/token-psa veraison/psadecoder")
MODULES+=("plugins/trustanchorstore-sqlite veraison/sqlitetastore")
MODULES+=("plugins/opapolicyengine veraison/opapolicyengine")
MODULES+=("plugins/sqliteendorsements github.com/veraison/sqliteendrosements")
MODULES+=("plugins/token-dice veraison/dicetoken")
MODULES+=("plugins/sqlitepolicy github.com/veraison/sqlitepolicy")
MODULES+=("tokenprocessor github.com/veraison/tokenprocessor")
MODULES+=("common github.com/veraison/common")
MODULES+=("verifier github.com/veraison/verifier")
MODULES+=("policy github.com/verason/policymanager")

for module in "${MODULES[@]}"
do
  dir=$(echo "${module}" | cut -d' ' -f1)
  mod=$(echo "${module}" | cut -d' ' -f2)

  echo ">> retrieving licenses [ ${mod} ]"
  ( cd "${dir}" && go-licenses csv ${mod} )
done
