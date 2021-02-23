# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * GOLINT - name of the linter package
# * GOLINT_ARGS - command line arguments for $(LINT) when run on the lint target
# * GOLINT_EXTRA_ARGS - command line arguments for $(LINT) when run on the lint-extra target
# targets:
# * lint  - run source code linter
# * lint-extra  - run source code linter

GOLINT ?= golangci-lint
# default linters (see ../.golang-ci.yml)
GOLINT_ARGS ?= run --timeout=3m
# optional linters (not required to pass)
GOLINT_EXTRA_ARGS ?= run --timeout=3m --issues-exit-code=0 -E dupl -E gocritic -E gosimple -E lll -E prealloc

.PHONY: lint
lint: ; $(GOLINT) $(GOLINT_ARGS)

.PHONY: lint-extra
lint-extra: ; $(GOLINT) $(GOLINT_EXTRA_ARGS)
