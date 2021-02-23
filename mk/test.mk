# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOPKG - name of the package to test
# * TEST_ARGS - command line arguments to go test
# targets:
# * test - run $(GOPKG) unit tests

TEST_ARGS ?= -v -cover -race

.PHONY: test
test: ; go test $(TEST_ARGS) $(GOPKG)
