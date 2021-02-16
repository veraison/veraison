# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0

SUBDIR := plugins
SUBDIR += common
SUBDIR += endorsement
SUBDIR += policy
SUBDIR += tokenprocessor
SUBDIR += verifier

include mk/subdir.mk
