# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * CLEANFILES - files to remove on clean
# targets:
# * all   - no-op
# * clean - remove $(CLEANFILES)

.PHONY: all
all:

.PHONY: clean
clean:
ifdef CLEANFILES
	$(RM) $(CLEANFILES)
endif
