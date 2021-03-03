package iat

evidence := input.evidence
endorsements := input.endorsements

default allow = false
allow {
	hardware_verified
	software_components_valid
	all_sw_components_matched
}

#  hardware id not specified
hardware_verified {
	in_hw := object.get(evidence, "HwVersion", "")
	in_hw == ""
}

# or

# if hardware id is specified, it must match the one returned for the verif scheme.
hardware_verified {
	in_hw := object.get(evidence, "HwVersion", "")
	in_hw == endorsements.hardware_id
}

default software_components_valid = false
software_components_valid {
	count(endorsements.software_components) > 0
}

default all_sw_components_matched = false
all_sw_components_matched {
	# the total number of evidence components matched to scheme components is equal
	# to the number of scheme components (i.e. all registered scheme components
	# have been mached in the evidence).
	count([x | sw_component_match[x]]) == count(endorsements.software_components)
}

# return the mached index inside scheme software for the evidence component at index i
sw_component_match[i] = j {
	some j
	in_comp := evidence.SwComponents[i]
	endorsements_comp := endorsements.software_components[j]

	in_comp.MeasurementType == endorsements_comp.sw_component_type
	in_comp.SignerID == endorsements_comp.signer_id
	in_comp.Version == endorsements_comp.sw_component_version
}
