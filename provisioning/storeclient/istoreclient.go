package storeclient

import (
	"github.com/veraison/endorsement"
)

type IStoreClient interface {
	endorsement.StoreClient
	endorsement.ProvisionerClient
}
