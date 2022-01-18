package decoder

type IDecoderManager interface {
	Init(dir string) error
	Close() error
	Dispatch(mediaType string, data []byte) (*EndorsementDecoderResponse, error)
	IsSupportedMediaType(mediaType string) bool
	SupportedMediaTypes() string
}
