// Copyright 2022 Contributors to the Veraison project.
// SPDX-License-Identifier: Apache-2.0
package storeclient

import (
	"github.com/veraison/common"
)

type IStoreClient interface {
	common.ProvisionerClient
	common.StoreClient
}
