package decoder

type Params map[string]interface{}

type IDecoder interface {
	Init(params Params) error
	Close() error
	GetName() string
	GetSupportedMediaTypes() []string
	Decode([]byte) (*EndorsementDecoderResponse, error)
}
