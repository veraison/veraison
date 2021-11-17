# Endorsement Provisioning Interface

The API described here are can be used to provision 
Endorsements into a Verifier. The API are agnostic with regards to 
the specific data model used to transport Endorsements. HTTP Content negotiation is used to determine the precise message structure and format of the information exchanged between the clients and the server. 
One specific example of information exchange using
[Concise Reference Integrity Manifest](https://datatracker.ietf.org/doc/draft-birkholz-rats-corim/) is given below.

## Provisioning API

The provisioning API allows authorised supply chain actors to communicate reference and endorsed values,
verification and identity key material, as well as other kinds of endorsements to Veraison. The supported 
format is CoRIM.

* To initiate a provisioning session, a client `POST`'s the CoRIM containing the endorsements to be provisioned to the `/submit` URL.
* If the transaction completes synchronously, a `200` response is returned to the client to indicate a
  successful submission of the posted CoRIM.
* For large submissions with multiple CoRIMs, the provisioning processing may happen in background. In this case,
  the Server returns a `201` response, with a `Location` header pointing to a "session" resource.  This "session" resource is only active for a given amount of time. Client can poll the resource to query the status of the submission request until the session completes.
* A session can be in one of the following states: `processing`, `complete`, or `failed`.
* The resource has a defined "time to live": upon its expiry, the resource is garbage collected.
* Client can continue periodic polling of the session uri, until the processing is either `complete` or `failed`.


## Synchronous submission

```
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ <------------------------|             |
            200 (OK)         '-------------' 

```

* Client submits the endorsement provisioning request
* Server responds with response code `200` indicating successful submission. 
  The transaction is complete
  
```
>> Request:
  POST /endorsement-provisioning/v1/submit
  Host: veraison.example
  Content-Type: application/rim+cbor

  ...CoRIM as binary data...

<< Response:
  HTTP/1.1 200 OK
```

## Asynchronous submission

```
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ \ <----------------------|             |
   \ \      201 ( Session )  '-------------' 
    \ \                             |
     \ \                            V
      \ \  (2) GET           .-------------.
       \ '------------------>| /session/01 |
        \<-------------------|             |
             200 (OK)        '-------------'
             status= "complete"
```

* Client submits the endorsement provisioning request
* Server responds with response code `201` indicating that the request has been accepted and will be processed asynchronously
* Server returns a time-bound session resource in the `Location` header. The resource can be polled at regular intervals to check the progress of the submission, until the processing is complete (either successfully or with a failure)

```
>> Request:
  POST /endorsement-provisioning/v1/submit
  Host: veraison.example
  Content-Type: application/rim+cbor

...CoRIM as binary data...
  
<< Response:
  HTTP/1.1 201 Created
  Content-Type: application/provisioning-session+json
  Location: /endorsement-provisioning/v1/session/1234567890

  {
    "status": "processing",
    "expiry": "2030-10-12T07:20:50.52Z"
  }
```

### Polling the session resource
```
>> Request:
  GET /endorsement-provisioning/v1/session/1234567890
  Host: veraison.example
  Accept: application/provisioning-session+json

<< Response:
  HTTP/1.1 200 OK
  Content-Type: application/provisioning-session+json

  {
    "status": "complete",
    "expiry": "2030-10-12T07:20:50.52Z"
  }
```
