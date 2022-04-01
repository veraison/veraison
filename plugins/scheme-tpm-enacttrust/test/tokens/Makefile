# Copyright 2022 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

all: ../cmd/gen-token/gen-token  tokens
.PHONY: all

TOKENS := $(addsuffix .token, $(basename $(wildcard *.json)))

CLEANFILES += $(TOKENS)

../cmd/gen-token/gen-token:
	$(MAKE) -C ../cmd/gen-token/

%.token: %.json ../cmd/gen-token/gen-token
	../cmd/gen-token/gen-token -out $@ -key ../keys/$(basename $<).pem $<

tokens: $(wildcard *.json)
	$(foreach token, $(TOKENS), $(MAKE) $(token))

test: ;
.PHONY: test

lint lint-extra: ;
.PHONY: lint lint-extra


include ../../../../mk/common.mk
include ../../../../mk/pkg.mk
