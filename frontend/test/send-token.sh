#!/bin/bash

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

sessionPath=$(curl -s -i -X POST http://localhost:8080/challenge-response/v1/newSession?nonceSize=32 | grep Location | cut -f2 -d: | tr -d ' \r')

curl -i -H "Content-Type: application/psa-attestation-token" -X POST http://localhost:8080$sessionPath -d "@$SCRIPT_DIR/psa-token.cbor"
