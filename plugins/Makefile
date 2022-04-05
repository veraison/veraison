# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := opapolicyengine
SUBDIR += sqliteendorsements
SUBDIR += sqlitepolicy
SUBDIR += scheme-dice
SUBDIR += scheme-psa
SUBDIR += scheme-tpm-enacttrust


ifeq ($(MAKECMDGOALS),test)
ifdef CI_PIPELINE

# At present, the policy engine built on OPA has no tests (see
# https://github.com/veraison/veraison/issues/17). We need to remove
# it from the CI testing because it messes up the coverage collection filter.
SUBDIR := $(filter-out opapolicyengine,$(SUBDIR))

endif # in CI pipeline
endif # test

include ../mk/subdir.mk
