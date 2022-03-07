// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package storeclient

import (
	"github.com/veraison/endorsement"
)

type IStoreClient interface {
	endorsement.StoreClient
	endorsement.ProvisionerClient
}
