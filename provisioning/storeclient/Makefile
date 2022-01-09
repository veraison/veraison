# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

GOPKG := github.com/veraison/veraison/provisioning/storeclient

all-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../common protogen

include ../../mk/common.mk
include ../../mk/pkg.mk
include ../../mk/lint.mk
include ../../mk/test.mk
