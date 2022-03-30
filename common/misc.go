// Copyright 2021 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"strings"
)

// Canonize converts the specified name into a "canonical" version by
// converting it to all lower case and replaces all hyphens to underscores.
// This is intended to be used on user-provided specifiers to allow for some
// flexibility therein.
func Canonize(name string) string {
	name = strings.ToUpper(name)
	name = strings.ReplaceAll(name, "-", "_")
	return name
}
