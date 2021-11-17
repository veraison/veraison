# Interaction models

The APIs described here are meant to be instantiations of the abstract
protocols described in [Reference Interaction Models for Remote Attestation
Procedures](https://datatracker.ietf.org/doc/draft-ietf-rats-reference-interaction-models/).

## Challenge/Response

Each Challenge-Response session is associated to its own resource, with the
following attributes:

* The session nonce;
* An expiry date after which the session is garbage collected;
* The accepted MIME types for Evidence to submit;
* The session state (`waiting` -> `processing` -> `complete` | `failed`)
* The submitted Evidence;
* The produced Attestation Result.

The resource is created in response to a client `POST` (1).  Subsequently, the
client interacts with its session resource by `POST`ing Evidence to verify (2),
and possibly polling it until the Attestation Result pops up (3).  In (2), the
server may decide to reply synchronously by including the Attestation Result
directly in the response.  In such case, (3) is not necessary.  The optional
cleanup step in (4) allows a client to explicitly destroy the session resource.
In any case the resource is garbage collected at any point in time after the
session expiry has elapsed.

```
 o        (1) POST           .-------------.
/|\ ------------------------>| /newSession |
/ \ \                        '------+------'
 \ \ \                              |
  \ \ \                             V
   \ \ \  (2) POST Evidence  .-------------.
    \ \ '------------------->| /session/01 |
     \ \                     '-------------'
      \ \                        ^  ^
       \ \    (3) GET            |  |
        \ '----------------------'  |
         \     (4) DELETE           |
          '-------------------------'
```

### Session setup

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

### Asynchronous verification

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

### Synchronous verification

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
