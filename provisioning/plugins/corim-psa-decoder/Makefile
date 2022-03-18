# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PLUGIN := ../bin/veraison-provisining-decoder-corim-psa
GOPKG := github.com/veraison/veraison/provisioning/plugins/corim-psa
SRCS := $(wildcard *.go)

all-hook-pre all-test-pre all-lint-pre:
	$(MAKE) -C ../../../common protogen
	$(MAKE) -C ../../decoder protogen

include ../../../mk/common.mk
include ../../../mk/plugin.mk
include ../../../mk/lint.mk
include ../../../mk/test.mk
