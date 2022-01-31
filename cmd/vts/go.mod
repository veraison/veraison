module github.com/veraison/cmd/trustedservices

go 1.15

replace github.com/veraison/common => ../../common/

replace github.com/veraison/endorsement => ../../endorsement/

require (
	github.com/veraison/common v0.0.0
	github.com/veraison/endorsement v0.0.0
	google.golang.org/grpc v1.41.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
