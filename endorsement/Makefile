# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

.DEFAULT_GOAL := test

GOPKG := github.com/veraison/endorsement

PROTOSRCS := storeapi.proto
PROTOPATHS := ../common

CLEANFILES := *.pb.go

lint-hook-pre: ; $(MAKE) protogen

include ../mk/common.mk
include ../mk/pkg.mk
include ../mk/lint.mk
include ../mk/proto.mk
include ../mk/test.mk
