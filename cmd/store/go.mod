module github.com/veraison/cmd/store

go 1.15

replace github.com/veraison/common => ../../common/

replace github.com/veraison/endorsement => ../../endorsement/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/veraison/endorsement v0.0.0
	google.golang.org/grpc v1.41.0
)
