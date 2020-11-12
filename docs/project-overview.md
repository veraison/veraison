# Project VERAISON Overview

Project _VERAISON_: **VER**ific**A**t**I**on of atte**S**tati**ON**

## Philosophy
The focus of this project is to produce components that are useful in a variety of deployments.
To that end there is a great focus on flexibility, allowing the use of plugins or policy modules to specialise the general flows that the components support.

## Get in Touch
If the following resonates with any of your plans, if you would like to contribute or if you would like to discuss the project approach in any way, please get in touch via the [Veraison Zulip channel](https://veraison.zulipchat.com)

## Reference Implementations
### Tokens
In order that the general logic within the components performs the correct set of operations to provide support for the maximum set of deployments, many real world attestation solutions and potential extrapolations of their use have been considered. Developing understanding and modelling the necessary flows for their verification process has driven the Veraison design.
To ensure that the code backs up the design, reference support for multiple tokens will be developed in the project. The current targets for this support are:
* EAT Tokens, with the ARM PSA Profile being a specific test set. 
* DICE Tokens. NB: the current delivery format for the 'token', being the full DICE Certificate chain is still under consultation. If you're planning any such protocol implementations, we would be very interested in discussing them.

### Deployments
To ensure that the core components are structured correctly for use as a service, the project will build reference deployments that can perform token verifications. To help ensure abstractions for the components over deployment environments are sufficiently general, multiple reference deployments will be built. They are:
* A deployment for self hosting
* A deployment targeting a hyperscaler PAAS environments
Further details on the above will be available within specific documentation on the reference deployment. (ToDo) 

## Scope 
The following is the intended scope of the Veraison Project. This is being listed to help others understand the intended functionality. If you're considering adopting or contributing to Veraison and you think there's something missing for the core of an Attestation Verification service, let us know via the Zulip channel. The following list should not be treated as a backlog, work items will be recorded elsewhere. 
### Scope - Verification
* An API to which an Attestation token, containing evidence claims, can be submitted. The API can either produce a simple boolean 'approved' response or a complex set of Attestation Results (verified claim summary). The results will be in a (signed) JSON format.
* The implementation of the API will be extensible to new formats & scenarios (e.g. encrypted tokens) by plugins
* The Appraisal Policy for the submitted evidence can be specified as a policy using the OPA (Open Policy Agent) language 'rego'. A plugin point for code, to support complex appraisal, will also be available.
* The Policy can access Reference Values or Endorsements required to support the Appraisal by specifying queries
* The API will record basic metrics for integration into deployment operational schemes
* In addition to the direct Verification of the supplied evidence, further policy can be deployed to add 'Derived Claims' to the Attestation results.

### Scope - Provisioning
Some of the more complex problems in the implementation of Verification Services arise from the requirements of Appraisal Policy to access Reference Values for Evidence. These values must be supplied by a trusted supply chain and kept securely by the Service. The Veraison project supports a Provisioning module that allows Reference Values (& Endorsements) to be presented by the supply chain and stored within the service. The actual storage medium is abstracted from the storage model. It has been determined that the most appropriate form of storage to support queries by the validator is a database that support graph structures & queries.
Given the above, the scope of the Provisioning work is:
* API to allow provisioning of Reference Values / Endorsements from supply chain. Secured provisioning actions will be validated.
* Bulk provisioning
* Upgrade Relationships for Firmware components
* Targeted Object Revocation
* Abstracted Storage
* Mapped to underlying Data Model to support Verification operations
* Support for some classes of remotely stored Reference Values (e.g. Project Trillian based Firmware Transparency)
* Multi Tenancy data separation
* AuthN & AuthZ access control for API operations & data separation
* Audit Trail for Provisioning operations
* Potential to extend operations for multi region support

### Scope - Reference Deployments
Reference deployments show that the APIs can be practically applied. The intent is that the Reference Deployments implemented can be a starting point for production deployments.
* Self Hosted deployment e.g. servers in a corporate data centre
* PAAS Hosted deployment e.g. AWS / Azure / GCP (other hyperscalers are available)
* APIs will be hosted and externally available
* Storage will be provided to support the data model
* Scaling out of API compute units in response to load
* Simple AuthN schemes
* Metric & Audit storage

The following are currently out of scope for the Reference Deployments:
* Full devops / CD work
* SLA monitoring
* Billing


## What's out of Scope for the Project
It is not intended to look at other aspects of Attestation e.g.
* Unification of Attestation Token formats
* Normalising Relying Party Attestation request API
* Common Attestation protocols


## Programming Language
GoLang is used as the implementation language for the project. This choice reflected the need for an efficient, memory managed language with good interoperability to other systems.

## Consuming Veraison
Veraison will achieve its aims if the componentry is used to build verification services. There are several ways we anticipate the components being incorporated into such services. We're always interested in discussing alternative integrations if these don't quite match project requirements. 
* As code modules - the modular nature of development of course supports this but may miss out on benefits of running as part of a wider system. 
* As 'compute units'. A compute unit is a deployable set of code that will present an API used for verification or infrastructure purposes. Compute units combine to build into deployable services, being units of scalability.
* Built as extensions from the Reference Deployments. The reference deployments built by the project may not be production ready, it's impossible to predict everyone's business / ops / dev integration requirements, but the intent of building them is that they will form a practical basis for real deployment. 

## Terminology
The Veraison Project will be using Terminology relating to Attestation as defined by the IETF Remote Attestation (RATS) working group. See the Terminology section from the [RATS Architecture document](https://tools.ietf.org/html/draft-ietf-rats-architecture-07)
