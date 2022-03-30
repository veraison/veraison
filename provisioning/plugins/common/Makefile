# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

all: all-hook-pre
clean:

all-hook-pre all-test-pre all-lint-pre: ; $(MAKE) -C ../../decoder protogen

include ../../../mk/common.mk
include ../../../mk/lint.mk
include ../../../mk/test.mk
