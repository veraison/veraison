# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PLUGIN := ../bin/opa-policy-engine
GOPKG := veraison/opapolicyengine
SRCS := main.go

all-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../common protogen

include ../../mk/common.mk
include ../../mk/plugin.mk
include ../../mk/lint.mk
include ../../mk/test.mk
