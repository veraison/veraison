# Endorsement Provisioning Interface

The APIs described here are meant to illustrate the provisioning of 
Endorsements based on RATS architecture. The API's defined here are not constrained 
by any specific data model. HTTP Content negotiation will be used to negotiate the precise message structure and format of the information exchanged between the clients and the server. 
One specific example of information exchange using the Data Model as
illustrated in  [Concise Reference Integrity Manifest](https://datatracker.ietf.org/doc/draft-birkholz-rats-corim/) is given below.

## Provisioning API

Provisioning API allows authorised supply chain actors to communicate reference and endorsed values,
verification and identity key material, as well as other kinds of endorsements to Veraison. The supported 
format is CoRIM.

* To initiate a provisioning session, a client `POST`'s the CoRIM (or a batch of CoRIMs) containing endorsements  to /`submit` URL;
* If the transaction completes synchronously and a `(200)`response is returned to the client, it indicates a
  successful submission of the posted CoRIM (or a batch of CoRIM). The ongoing transaction is now complete.
* For large submission with multiple CoRIMs, the provisioning processing may happen in background. In this case,
  Client is returned a `(201)` response, with a link header populated with a session resource. Resource is only active for a given amount of time. Client will use the resource to poll the request till the session is active.
* The session progresses as (`processing` -> `complete` | `failed`)
* The resource has a defined `time to live` upon its expiry automatic garbage collection is performed.
* Client can continue periodic polling of the session uri, until the processing is either `complete` or `failed`


## Synchronous submission

```
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ \ <----------------------|             |
 \ \ \      200 (OK)         '-------------' 

```

## Asynchronous submission

```
 o        (1) POST           .-------------.
/|\ ------------------------>| /submit     |
/ \ \ <----------------------|             |
 \ \ \      201 ( Session )  '-------------' 
  \ \ \                             |
   \ \ \                            V
    \ \ \  (2) GET            .-------------.
     \ \ '------------------->| /session/01 |
      \ \<--------------------|             |
             200 (OK)         '-------------'
             status= "completed"
```

### Synchronous POST

* Client submits the endorsement provisioning request
* Server responds with response code `200` indicating successful submission. 
  The transation is now complete
```
>> Request:
  POST /endorsement-provisioning/v1/submit/
  Host: veraison.example
  Content-Type: application/rim+cbor

  {
    "payload": "56789abcdesgjgwasdgh"
  }

<< Response:
  HTTP/1.1 200 OK
```

### Asynchronous POST

* Client submits the endorsement provisioning request
* Server responds with response code `201` indicating that the request will be processed asynchronously.
* Server returns a time bound resource in response header. The resource will be used to poll the original   request till the time it is complete (successful/failed) or the session expires.

```
>> Request:
  POST /endorsement-provisioning/v1/submit/
  Host: veraison.example
  Content-Type: application/rim+cbor

  {
    "payload": "489765abcdesgfdwdunisdufbiubwjndisjd"
  }
  
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
    "status": "completed",
    "expiry": "2030-10-12T07:20:50.52Z"
  }
```