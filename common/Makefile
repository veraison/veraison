# Copyright 2021-2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/common

PROTOSRCS := result.proto
PROTOSRCS += attestation_format.proto
PROTOSRCS += param.proto
PROTOSRCS += token.proto
PROTOSRCS += evidence.proto
PROTOSRCS += attestation.proto
PROTOSRCS += vts.proto

SUBDIR := test/plugin/

test-hook-pre:
	$(MAKE) -C test/plugin/

include ../mk/common.mk
include ../mk/proto.mk
include ../mk/pkg.mk
include ../mk/lint.mk
include ../mk/test.mk
include ../mk/subdir.mk
