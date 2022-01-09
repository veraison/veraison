# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

PROTOSRCS := decoder_response.proto

PROTOPATHS := ../../common
PROTOPATHS += ../../endorsement

CLEANFILES += *.pb.go

all-hook-pre test-hook-pre lint-hook-pre:
	$(MAKE) -C ../../common protogen
	$(MAKE) -C ../../endorsement protogen
	$(MAKE) protogen

include ../../mk/common.mk
include ../../mk/pkg.mk
include ../../mk/lint.mk
include ../../mk/proto.mk
include ../../mk/test.mk
