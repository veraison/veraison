# Endorsement Provisioning Interface

The API described here are can be used to provision 
Endorsements into a Verifier. The API are agnostic with regards to 
the specific data model used to transport Endorsements. HTTP Content negotiation is used to determine the precise message structure and format of the information exchanged between the clients and the server. 
One specific example of information exchange using
[Concise Reference Integrity Manifest (CoRIM)](https://datatracker.ietf.org/doc/draft-birkholz-rats-corim/) is given below.

## Provisioning API

The provisioning API allows authorized supply chain actors to communicate reference and endorsed values,
verification and identity key material, as well as other kinds of endorsements to Veraison. The supported 
format is CoRIM.

* To initiate a provisioning session, a client `POST`'s the CoRIM containing the endorsements to be provisioned to the `/submit` URL.
* If the transaction completes synchronously, a `200` response is returned to the client to indicate the
  submission of the posted CoRIM has been fully processed.  The response body contains a "session" resource whose `status` field encodes the outcome of the submission (see below).
* The provisioning processing may also happen asynchronously, for example when submitting a large CoRIM. In this case,
  the server returns a `201` response, with a `Location` header pointing to a session resource that the client can regularly poll to monitor any change in the status of its request.
* A session starts in `processing` state and ends up in one of `success` or `failed`. When in `failed` state, the `failure-reason` field of the session resource contains details about the error condition.
* The session resource has a defined time to live: upon its expiry, the resource is garbage collected.  Alternatively, the client can dispose the session resource by issuing a `DELETE` to the resource URI.

## Synchronous submission

```text
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ <------------------------|             |
            200 (OK)         '-------------' 
            { "status": ... }
```

* Client submits the endorsement provisioning request
* Server responds with response code `200` indicating processing is complete.
  The response body contains a session resource with a `status` indicating the outcome of the submission operation.

### Example of a successful submission

```text
>> Request:
  POST /endorsement-provisioning/v1/submit
  Host: veraison.example
  Content-Type: application/rim+cbor

  ...CoRIM as binary data...

<< Response:
  HTTP/1.1 200 OK
  Content-Type: application/vnd.veraison.provisioning-session+json

  {
    "status": "success",
    "expiry": "1970-01-01T00:00:00Z"
  }
```

### Example of a failed submission

```text
>> Request:
  POST /endorsement-provisioning/v1/submit
  Host: veraison.example
  Content-Type: application/rim+cbor

  ...CoRIM as binary data...

<< Response:
  HTTP/1.1 200 OK
  Content-Type: application/vnd.veraison.provisioning-session+json

  {
    "status": "failed",
    "failure-reason": "invalid signature",
    "expiry": "1970-01-01T00:00:00Z"
  }
```

## Asynchronous submission

```text
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ \ <----------------------|             |
   \ \      201 Created      '-------------'
    \ \     Location: /session/01   |
     \ \                            V
      \ \  (2) GET           .-------------.
       \ '------------------>| /session/01 |
        `<-------------------|             |
             200 OK          '-------------'
             { "status": ... }
```

* Client submits the endorsement provisioning request
* Server responds with response code `201` indicating that the request has been accepted and will be processed asynchronously
* Server returns the URI of a time-bound session resource in the `Location` header. The resource can be polled at regular intervals to check the progress of the submission, until the processing is complete (either successfully or with a failure)

### Example


```text
>> Request:
  POST /endorsement-provisioning/v1/submit
  Host: veraison.example
  Content-Type: application/rim+cbor

...CoRIM as binary data...
  
<< Response:
  HTTP/1.1 201 Created
  Content-Type: application/vnd.veraison.provisioning-session+json
  Location: /endorsement-provisioning/v1/session/1234567890

  {
    "status": "processing",
    "expiry": "2030-10-12T07:20:50.52Z"
  }
```

```text
>> Request:
  GET /endorsement-provisioning/v1/session/1234567890
  Host: veraison.example
  Accept: application/vnd.veraison.provisioning-session+json

<< Response:
  HTTP/1.1 200 OK
  Content-Type: application/vnd.veraison.provisioning-session+json

  {
    "status": "complete",
    "expiry": "2030-10-12T07:20:50.52Z"
  }
```
