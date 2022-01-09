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

_mocks: $(MOCK_FILES)
.PHONY: _mocks

test: test-hook-pre realtest
.PHONY: test

test-hook-pre:
.PHONY: test-hook-pre

realtest: _mocks ; go test $(TEST_ARGS) $(GOPKG)
.PHONY: realtest

CLEANFILES += $(MOCK_FILES)
