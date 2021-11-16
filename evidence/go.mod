module github.com/veraison/endorsement

go 1.15

replace github.com/veraison/common => ../common

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/veraison/common v0.0.0
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
)
