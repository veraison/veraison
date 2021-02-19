## Attestation Token use cases & assumptions
The following captures what is assumed about the nature of the attestation tokens received for verification.

A token is presented as a single 'blob' (byte stream).

Tokens are presented individually. If a batch mode submission is required, it is the responsibility of an API tier to pass them for verification one by one.

Encoding in transit is the responsibility of the API tier, the verification pipeline works on a byte stream.

The verification pipeline will receive some context information which will indicate or help to deduce the nature of the token and hence which plugins will be needed in the verification pipeline. In a dedicated deployment this may be hardwired, in an environment supporting multiple token types, this is likely to come from the API tier.

A token may have an encrypted body. Assuming the correct decryption key is available, the body will be decrypted and then passed into the pipeline in the same manner as an unencrypted token.

A presented token blob may actually contain multiple tokens, each of which will be handled e.g. deserialised and have signing checked, independently such that an aggregation of all evidence contained is extracted and presented for appraisal

Multiple token models include both concatenation (e.g a chain of certs packaged serially) or embedded (e.g. the value of a claim within one token may be a serialised independent token)

The set of endorsements / reference values required to appraise the evidence in a token is assumed to have been provisioned to the verifier service prior to the token being presented. i.e. the verifier cannot distinguish between a bad value in a claim and a 'missing' value.

The exception to the above might be where the R.V. is held in an external service which the verifier must consult. In this case, the prior rule applies to that external service.
