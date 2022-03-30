# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * SRCS       - plugin source files
# * PLUGIN     - plugin binary file
# * DEBUG      - set this to true to compile with debug symbols
# * CLEANFILES - any additional file to remove on clean
# targets:
# * all   - build $(PLUGIN) from $(SRCS) [DEFAULT]
# * clean - remove $(PLUGIN)

.DEFAULT_GOAL := all

ifndef PLUGIN
  $(error PLUGIN must be set when including plugin.mk)
endif

ifdef DEBUG
  DFLAGS := -gcflags='all=-N -l'
else
  DFLAGS :=
endif

$(PLUGIN): $(SRCS) ; go build $(DFLAGS) -o $(PLUGIN)

.PHONY: all
all: all-hook-pre realall

.PHONY: all-hook-pre
all-hook-pre:

.PHONY: realall
realall: $(PLUGIN)

.PHONY: clean
clean: ; $(RM) $(PLUGIN) $(CLEANFILES)
