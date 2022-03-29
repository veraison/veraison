#!/bin/bash

set -o pipefail
set -eux

T=${T?must be set in the environment to one of psa or tpm-enacttrust.}
B=${B?must be set in the environment to one of trustanchor or refvalue.}

CORIM_FILE=corim-${T}-${B}.cbor
CORIM_FILE=corim-${T}-${B}.cbor

if [ "${T}" == "psa" ]
then
	CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1"
elif [ "${T}" == "tpm-enacttrust" ];
then
	CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://enacttrust.com/veraison/1.0.0"
else
	echo "unknown type ${T}"
	exit 1
fi

curl --include \
	--data-binary "@${CORIM_FILE}" \
	--header "${CONTENT_TYPE}" \
	--header "Accept: application/vnd.veraison.provisioning-session+json" \
	--request POST \
	http://localhost:8888/endorsement-provisioning/v1/submit 
