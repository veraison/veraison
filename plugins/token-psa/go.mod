module veraison/psadecoder

go 1.15

replace github.com/veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/veraison/common v0.0.0
	github.com/veraison/psatoken v0.0.1
)
