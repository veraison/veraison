package decoder

type IDecoderManager interface {
	Init(dir string) error
	Term() error
	Dispatch(mediaType string, data []byte) (*EndorsementDecoderResponse, error)
	IsSupportedMediaType(mediaType string) bool
	SupportedMediaTypes() string
}
