.DEFAULT_GOAL := help

# To validate an OpenAPI module, append the file name (without the .yaml
# extension) to the API variable.  A "check_<YAML module>" target is
# automatically constructed and appended to the top level "check" goal.
API := endorsement-provisioning

CMD := openapi-spec-validator

ifeq (, $(shell which $(CMD)))
$(error "No openapi-spec-validator in $(PATH); To install it, run: 'pip install openapi-spec-validator'")
endif

define OPENAPI_CHECK_template
.PHONY: check_$(1)

check_$(1): ; @printf ">>> $(1): " && $(CMD) $(1).yaml

HELP += "check_$(1): check only the $(1) template"
endef

$(foreach a,$(API),$(eval $(call OPENAPI_CHECK_template,$(a))))

.PHONY: check
check: $(foreach t,$(API),check_$(t))

.PHONY: help
help:
	@printf "\nAvailable targets:\n"
	@printf "\tcheck: run the OpenAPI spec validator on all API templates\n\n"
	@printf "\t%s\n\n" $(HELP)
