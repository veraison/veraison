# Challenge/Response

## Session setup

* To start a new session and obtain a time-bounded nonce value:

```
>> Request:
  POST /challenge-response/v1/newSession?nonceSize=32
  Host: veraison.example
  Accept: application/rats-challenge-response-session+json

<< Response:
  HTTP/1.1 201 Created
  Content-Type: application/rats-challenge-response-session+json
  Location: https://veraison.example/challenge-response/v1/session/1234567890

  {
    "nonce": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=",
    "expiry": "2030-10-12T07:20:50.52Z",
    "accept": [
      "application/psa-attestation-token"
    ],
    "state": "waiting"
  }
```

## Asynchronous verification

* Submit evidence for this session:

```
>> Request:
  POST /challenge-response/v1/session/1234567890
  Host: veraison.example
  Accept: application/rats-challenge-response-session+json
  Content-Type: application/psa-attestation-token

  .....

<< Response:
  HTTP/1.1 202 Accepted
  Content-format: application/rats-challenge-response-session+json

  {
    "nonce": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=",
    "expiry": "2031-10-12T07:20:50.52Z",
    "accept": [
      "application/psa-attestation-token"
    ],
    "state": "processing",
    "evidence": {
      "type": "application/psa-attestation-token",
      "value": "eyJhbGciO...RfrKmTWk"
    }
  }
```

* Since we got back a 202, we need to poll for result:

```
>> Request:
  GET /challenge-response/v1/session/1234567890
  Host: veraison.example
  Accept: application/rats-challenge-response-session+json

<< Response:
  HTTP/1.1 200 OK
  Content-format: application/rats-challenge-response-session+json

  {
    "nonce": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=",
    "expiry": "2031-10-12T07:20:50.52Z",
    "accept": [
      "application/psa-attestation-token"
    ],
    "state": "complete",
    "evidence": {
      "type": "application/psa-attestation-token",
      "value": "eyJhbGciO...RfrKmTWk"
    },
    "result": {
      "is_valid": true,
      "claims": {
        // ...
      }
    }
  }
```

## Synchronous verification

* Submit evidence for this session and obtain the attestation result right away
  (200):

```
>> Request:
  POST /challenge-response/v1/session/1234567890
  Host: veraison.example
  Accept: application/rats-challenge-response-session+json
  Content-Type: application/psa-attestation-token

  .....

<< Response:
  HTTP/1.1 200 OK
  Content-format: application/rats-challenge-response-session+json

  {
    "nonce": "MTIzNDU2Nzg5MDEyMzQ1Njc4OTAxMjM0NTY3ODkwMTI=",
    "expiry": "2030-10-12T07:20:50.52Z",
    "accept": [
      "application/psa-attestation-token"
    ],
    "state": "complete",
    "evidence": {
      "type": "application/psa-attestation-token",
      "value": "eyJhbGciO...RfrKmTWk"
    },
    "result": {
      "is_valid": true,
      "claims": {
        // ...
      }
    }
  }
```
