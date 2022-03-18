# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOPKG - name of the package to test
# * TEST_ARGS - command line arguments to go test
# * INTERFACES - interface files to be mocked
# * MOCKPKG - name of the mock package
# targets:
# * test - run $(GOPKG) unit tests

TEST_ARGS ?= -v -cover -race
MOCKGEN := $(shell go env GOPATH)/bin/mockgen

define MOCK_template
mock_$(1): $(1)
	$$(MOCKGEN) -source=$$< -destination=mocks/$$$$(basename $$@) -package=$$(MOCKPKG)
endef

$(foreach m,$(INTERFACES),$(eval $(call MOCK_template,$(m))))
MOCK_FILES := $(foreach m,$(INTERFACES),$(join mock_,$(m)))

THIS_DIR := $(dir $(abspath $(lastword $(MAKEFILE_LIST))))

_mocks: $(MOCK_FILES)
.PHONY: _mocks

test: test-hook-pre realtest checkcopyrights
.PHONY: test

test-hook-pre:
.PHONY: test-hook-pre

realtest: _mocks ; go test $(TEST_ARGS) $(GOPKG)
.PHONY: realtest

COPYRIGHT_FLAGS :=
ifdef CI_PIPELINE
	COPYRIGHT_FLAGS += --no-year-check
endif

checkcopyrights: ; python3 $(THIS_DIR)../scripts/check-copyright $(COPYRIGHT_FLAGS) .
.PHONY: checkcopyrights

CLEANFILES += $(MOCK_FILES)
