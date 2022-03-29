module veraison/veraison/scheme/tpm-enacttrust

go 1.15

replace github.com/veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.4.3
	github.com/veraison/common v0.0.0-00010101000000-000000000000
	google.golang.org/protobuf v1.28.0
)
