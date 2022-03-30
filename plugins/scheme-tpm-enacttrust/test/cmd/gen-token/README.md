## Token generation

`gen-token` is used to generate test attestation tokens from JSON descriptions
of the payload. This is build and used when running the tests.

## Token Format

An attestation token has the following format:

    TPMS_ATTEST_LEN||TPMS_ATTEST||TPMS_ATTEST_SIGNATURE

where

`TPMS_ATTEST_LEN` is a uint16 values containing the length (in bytes) of the
following `TPMS_ATTEST` structure. `TPMS_ATTEST` is structured according to
[TPM 2.0
Specification](https://trustedcomputinggroup.org/wp-content/uploads/TCG_TPM2_r1p59_Part2_Structures_pub.pdf).
`TPMS_ATTEST_SIGNATURE` is the ECDSA signature of the `TPMS_ATTEST` structure.

## Token Payload Description Format

`gen-token` takes a path to a file containing the description of the payload
of the token. This description is a JSON object in the following form:

```
{
	"node-id": "7df7714e-aa04-4638-bcbf-434b1dd720f1",
	"firmware": 7,
	"pcrs": [1, 2, 3, 4],
	"algorithm": 4,
	"digest": "h0KPxSKAPTEGXnvOPPA/5HUJZjHl4Hu9eg/eYMTPJcc="
}
```

note: "signer" and "digest" are decoded as `[]byte`, and so, as per
[`encoding/json` marshalling rules](https://pkg.go.dev/encoding/json#Marshal),
should represented as the base64 encoding of their actual values.

## Key generation

`gen-key.sh` is used to generate EC P-256 keys that `gen-token` uses to create
the signature part of of the token. These keys are already pre-generated inside
../../vectors/keys/, and the script is not utilized when running the tests.
