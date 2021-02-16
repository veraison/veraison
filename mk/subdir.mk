# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * SUBDIR  - a list of subdirectories that should be built as well.
#             each of the targets will execute the same target in the
#             subdirectories.
# targets:
# * any target (optionally with -pre and -post suffix)

ifndef SUBDIR
  $(error SUBDIR must be set when including subdir.mk)
endif

.DEFAULT_GOAL := all
MAKECMDGOALS ?= $(.DEFAULT_GOAL)

# all targets (plain and hooks)
G = $(foreach T,$(MAKECMDGOALS),$(T)-pre $(addsuffix .$(T),$(SUBDIR)) $(T)-post)

# the cartesian product between MAKECMDGOALS and SUBDIR sets
G_PLAIN = $(filter-out %-pre %-post, $(G))

# hook'd MAKECMDGOALS
G_HOOK = $(filter %-pre %-post, $(G))

.PHONY: $(G)

$(MAKECMDGOALS): $(G)

# empty hooks (caller may override them)
$(G_HOOK):

# plain hooks (e.g. a_subdir.all, another_subdir.depend, ...)
$(G_PLAIN):
	@$(MAKE) -C $(basename $@) $(patsubst .%,%,$(suffix $@))
