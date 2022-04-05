# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := good

include

.DEFAULT_GOAL := build

.PHONY: build
build:
	mkdir -p bin/
	go build -o bin/good good/main.go

.PHONY: test
test:

.PHONY: protogen
protogen:

.PHONY: test
test:

CLEANFILES += bin/good

include ../../../mk/common.mk
include ../../../mk/pkg.mk
