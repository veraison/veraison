package common

import cbor "github.com/fxamacker/cbor/v2"

var (
	em, emError = initCBOREncMode()
	dm, dmError = initCBORDecMode()
)

func initCBOREncMode() (en cbor.EncMode, err error) {
	encOpt := cbor.EncOptions{
		Sort:          cbor.SortCanonical,
		IndefLength:   cbor.IndefLengthForbidden,
		TimeTag:       cbor.EncTagRequired,
		ShortestFloat: cbor.ShortestFloat16,
	}
	return encOpt.EncMode()
}

func initCBORDecMode() (dm cbor.DecMode, err error) {
	decOpt := cbor.DecOptions{
		IndefLength: cbor.IndefLengthForbidden,
	}
	return decOpt.DecMode()
}

func init() {
	if emError != nil {
		panic(emError)
	}
	if dmError != nil {
		panic(dmError)
	}
}
