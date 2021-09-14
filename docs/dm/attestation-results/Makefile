.DEFAULT_GOAL := check

SHELL := /bin/bash

%.cbor: %.diag ; diag2cbor.rb $< > $@

include tools.mk

check:: check-json check-json-examples
check:: check-cbor check-cbor-examples
check:: check-veraison-cbor check-veraison-cbor-examples
check:: check-veraison-json check-veraison-json-examples

# $1: label
# $2: cddl fragments
# $3: (optional) diag test files
define cddl_check_template

check-$(1): $(1)-autogen.cddl
	$$(cddl) $$< g 1 | $$(diag2diag) -e

.PHONY: check-$(1)

$(1)-autogen.cddl: $(2)
	for f in $$^ ; do ( grep -v '^;' $$$$f ; echo ) ; done > $$@

CLEANFILES += $(1)-autogen.cddl

check-$(1)-examples: $(1)-autogen.cddl $(3:.diag=.cbor)
	@for f in $(3:.diag=.cbor); do \
		echo ">> validating $$$$f" ; \
		$$(cddl) $$< validate $$$$f || exit 1; \
	done

.PHONY: check-$(1)-examples

CLEANFILES += $(3:.diag=.cbor)

endef # cddl_check_template

JSON_FRAGS := attestation-results.cddl
JSON_FRAGS += json-labels.cddl
JSON_FRAGS += generic-non-empty.cddl

JSON_EXAMPLES := $(wildcard examples/json-*.diag)

$(eval $(call cddl_check_template,json,$(JSON_FRAGS),$(JSON_EXAMPLES)))

CBOR_FRAGS := attestation-results.cddl
CBOR_FRAGS += cbor-labels.cddl
CBOR_FRAGS += generic-non-empty.cddl

CBOR_EXAMPLES := $(wildcard examples/cbor-*.diag)

$(eval $(call cddl_check_template,cbor,$(CBOR_FRAGS),$(CBOR_EXAMPLES)))

VERAISON_CBOR_FRAGS := attestation-results.cddl
VERAISON_CBOR_FRAGS += cbor-labels.cddl
VERAISON_CBOR_FRAGS += generic-non-empty.cddl
VERAISON_CBOR_FRAGS += veraison.cddl
VERAISON_CBOR_FRAGS += veraison-cbor-labels.cddl

VERAISON_CBOR_EXAMPLES := $(wildcard examples/veraison-cbor-*.diag)

$(eval $(call cddl_check_template,veraison-cbor,$(VERAISON_CBOR_FRAGS),$(VERAISON_CBOR_EXAMPLES)))

VERAISON_JSON_FRAGS := attestation-results.cddl
VERAISON_JSON_FRAGS += json-labels.cddl
VERAISON_JSON_FRAGS += generic-non-empty.cddl
VERAISON_JSON_FRAGS += veraison.cddl
VERAISON_JSON_FRAGS += veraison-json-labels.cddl

VERAISON_JSON_EXAMPLES := $(wildcard examples/veraison-json-*.diag)

$(eval $(call cddl_check_template,veraison-json,$(VERAISON_JSON_FRAGS),$(VERAISON_JSON_EXAMPLES)))

clean: ; $(RM) $(CLEANFILES)
