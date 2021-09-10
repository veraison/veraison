# Device and Supply Chain modelling

## Intro
This document sets out the accumulated knowledge and / or assumptions about:
+ the evidence available within a device which can be used for attestation
+ models for firmware release lifecycles 
+ observations on how multiple products may be released from a single source and the implications for provisioning data
Why are these significant? The models here have driven the design of the Veraison system, in particular the way supply chain data is provisioned into the system and the Endorsement store is constructed to allow verification queries to access reference values to be used for evidence appraisal. Those design conclusions can be found in other documents.

The Veraison project welcomes feedback from real world examples as to the accuracy of the models and assumptions used here. The project will be seeking confirmation on these models from industry partners. If you have views on these topics, please provide relevant commentary via a github issue or start a conversation on the Zulip channel.

## Device evidence classification
The evidence within an attestation report can be categorised into the following classes. These classes are not an exhaustive set of all types of evidence that may be found within a report but are those necessary for correlating evidence with endorsement / reference value queries.

### Trust Anchor
All devices are ASSUMED to contain a Trust Anchor (TA). This is normally provisioned into the device at manufacturing time but may also be provisioned to the device on enrolling into a management structure. The TA takes the form of a key or a seed from which an Attestation signing key can be derived. The attestation report will contain information, referred to here as a Trust Anchor ID (TA-ID), which a Verifier can use to locate an endorsement providing information on the public value of the relevant key to use to validate the signature on a report. The Trust Anchor may have unique-to-device or group scope (i.e. multiple devices contain the same trust anchor). The latter case has been used as one solution for privacy protection use cases. 

### Hardware Group
A device MAY contain a Hardware Group ID that is presented in attestation evidence. This is an identifier, provisioned during manufacturing, which can be used to classify the SoC used within the device and will be the same for any devices built using that SoC. Although termed as a hardware ID, in fact this identifier will classify the immutable Root of Trust, being the un-modifiable and un-measured initial BIOS as well as the specific SoC hardware within the SoC. The identifier is a public value and has no security value, but aids classification for a verifier. If no Hardware Group ID is included within the attestation evidence then an equivalent value can be provisioned into the verifier as a correlation to the Trust Anchor ID.

### Firmware
A device is ASSUMED to produce evidence allowing the set of loaded firmware components to be identified. In some cases the device may summarise all loaded components into a single version number. In other cases a full set of measurements and metadata may be be made available to identify the set of firmware components loaded.

### Workload
An attestation report may include a series of 

### Other Evidence
Other evidence may be contained within the Attestation Report that does not fall into the above categories. This evidence may need to be referenced as part of an evidence appraisal policy, but no additional classification for a group of evidence is known at the moment to affect the modelling of Endorsements for the project.

Note: attestation reports may contain additional complexities to modelling evidence, such as having multiple attesters within a device or receiving an encrypted report. However, it is expected that these can be handled as precursors to report verification by deconstruction into core parts in a pre-verification pass.

Diagram: relevant groupings for evidence

![Groupings](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/veraison/veraison/main/docs/diags/tokens-model.puml)

Diagram: Known Tokens

![Diagram: Mapping known attestation tokens to the evidence classifications](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/veraison/veraison/main/docs/diags/known-models.puml)


## Firmware Lifecycles
Where an attestation report contains only a firmware version, no additional data is required from the supply chain and whether the value provided is acceptible is a matter for the evidence appraisal policy.

The inclusion of measurements requires that the supply chain provisions information to the verifier about released firmware items such that the verifier can provide a range of valid references to an appraisal policy. 
At the same time, correlations can be established between firmware components and other evidence categories. This can be used to identify policy appraisals such as 'this hardware requires a certain set of firmware components to be trustworthy'.

Correlations can also be established between different released versions of firmware releases as a component is upgraded or patched. This allows a policy to gain appraisal policies such as whether a device is running up to date firmware or whether a firmware component is marked as having some vulnerability. Note that a distinction is made between an Update as being more likely to involve multiple fixes or components released at one time while a patch is more likely to be only a single component change to address a specific issue. 

The project has considered multiple models for likely firmware component lifecycles over revised releases. These can be classified into the following categories:
Model 1:
Updates: Release all components for whether changed or not (unchanged components have the same file released as on previous occasion)
Patches: Release only the patched (changed) component

Model 2:
Updates: As model 1
Patches: Release all components, only patch components will be changed unpatched components have the same file released as on previous occasion 
 
Model 3: 
Updates & Patches: Always release all components; always rebuild all components whether changed or not.
Note that if a vendor is following practices such as from https://reproducible-builds.org/ then Models 2 & 3 are equivalent.

Metadata can be added at any time to any released component via the provisioning system.

Diagram: Firmware lifecycle models

 ![Diagram: Firmware lifecycle models](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/veraison/veraison/main/docs/diags/firmware-lifecycles.puml)
 

## Beware clashes from multiple products
The above classifications can be used to build a data model which will allow the relevant __and related__ reference values / endorsements to be available for evidence appraisal. However, if a manufacturer releases multiple products using the same components *and* tokens from these products are handled by the same verification service, the creation of the data model must take into account these distinct products to avoid false lifecycle graphs. This is necessary to avoid, say, the implication that a firmware component is not up to date because a new release is available, which applies only to a different product. The concept of a product identity can be either created from a combination of evidence fields or be a created entity that is correlated to the TA-ID. 

Diagram: Multi Product Models

![Diagram: Multi product models](http://www.plantuml.com/plantuml/proxy?src=https://raw.githubusercontent.com/veraison/veraison/main/docs/diags/mult-prods.puml)


