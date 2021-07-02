.DEFAULT_GOAL := check

SHELL := /bin/bash

%.cbor: %.diag ; diag2cbor.rb $< > $@

include tools.mk

check:: check-json check-json-examples
check:: check-cbor check-cbor-examples

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
		$$(cddl) $$< validate $$$$f ; \
	done

.PHONY: check-$(1)-examples

CLEANFILES += $(3:.diag=.cbor)

endef # cddl_check_template

JSON_FRAGS := attestation-results.cddl
JSON_FRAGS += json-labels.cddl

JSON_EXAMPLES := $(wildcard examples/json-*.diag)

$(eval $(call cddl_check_template,json,$(JSON_FRAGS),$(JSON_EXAMPLES)))

CBOR_FRAGS := attestation-results.cddl
CBOR_FRAGS += cbor-labels.cddl

CBOR_EXAMPLES := $(wildcard examples/cbor-*.diag)

$(eval $(call cddl_check_template,cbor,$(CBOR_FRAGS),$(CBOR_EXAMPLES)))

clean: ; $(RM) $(CLEANFILES)
