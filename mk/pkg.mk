# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOPKG - name of the go package
# * CLEANFILES - files to remove on clean
# targets:
# * all   - no-op
# * clean - remove $(CLEANFILES)
# * test  - run $(GOPKG) tests [DEFAULT]
# * lint  - run source code linter

.DEFAULT_GOAL := test

.PHONY: all
all:

.PHONY: test
test: ; go test -v -cover -race $(GOPKG)

.PHONY: lint
lint: ; golangci-lint run

.PHONY: clean
clean:
ifdef CLEANFILES
	$(RM) $(CLEANFILES)
endif
