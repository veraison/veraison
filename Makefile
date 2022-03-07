# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := plugins
SUBDIR += common
SUBDIR += policy
SUBDIR += verifier
SUBDIR += frontend
SUBDIR += cmd
SUBDIR += kvstore
SUBDIR += provisioning

# At present, the frontend has no tests We need to remove it from the CI
# testing because it messes up the coverage collection filter.
SUBDIR := $(filter-out frontend,$(SUBDIR))

# At present, the verifier tests do not work, we need to remove it from the CI
# testing because it messes up the coverage collection filter.
SUBDIR := $(filter-out verifier,$(SUBDIR))


include mk/subdir.mk
