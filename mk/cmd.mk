# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * CMD      - the name of the binary that is built
# * SRCS     - the source files that CMD is build from
# * CMD_DEPS - any extra dependency
#
# targets:
# * all      - build the binary and save it to $(CMD)
# * clean    - remove the generated binary

ifndef CMD
  $(error CMD must be set when including cmd.mk)
endif

$(CMD): $(SRCS) $(CMD_DEPS) ; go build -o $(CMD)

CLEANFILES += $(CMD)

.PHONY: realall
realall: $(CMD)

.PHONY: cmd-hook-pre
cmd-hook-pre:

.PHONY: all
all: cmd-hook-pre realall
