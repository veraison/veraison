#!/bin/bash

set -eu
set -o pipefail

CORIM_TEMPLATE=corimMini.json

COMID_TEMPLATES=
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustAKOne"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustGoldenOne"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustAKMult"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustBadInst"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustNoInst"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustMultDigest"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustGoldenTwo"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustNoDigest"
COMID_TEMPLATES="${COMID_TEMPLATES} ComidTpmEnactTrustAKBadInst"

TV_DOT_GO=${TV_DOT_GO?must be set in the environment.}

printf "package main\n\n" > ${TV_DOT_GO}

for t in ${COMID_TEMPLATES}
do
	cocli comid create -t ${t}.json
	cocli corim create -m ${t}.cbor -t ${CORIM_TEMPLATE} -o corim${t}.cbor
	echo "// automatically generated from $t.json" >> ${TV_DOT_GO}
	echo "var unsignedCorim${t} = "'`' >> ${TV_DOT_GO}
	cat corim${t}.cbor | xxd -p >> ${TV_DOT_GO}
	echo '`' >> ${TV_DOT_GO}
	gofmt -w ${TV_DOT_GO}
done
