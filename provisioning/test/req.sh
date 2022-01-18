CORIM_FILE=corim-psa-trustanchor.cbor
#CORIM_FILE=corim-psa-refvalue.cbor

#CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://arm.com/psa/iot/1"
CONTENT_TYPE="Content-Type: application/corim-unsigned+cbor; profile=http://enacttrust.com/veraison/1.0.0"


curl --include \
	--data-binary "@${CORIM_FILE}" \
	--header "${CONTENT_TYPE}" \
	--header "Accept: application/vnd.veraison.provisioning-session+json" \
	--request POST \
	http://localhost:8888/endorsement-provisioning/v1/submit 
