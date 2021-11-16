module github.com/veraison/cmd/trustedservices

go 1.15

replace github.com/veraison/common => ../../common/

replace github.com/veraison/endorsement => ../../endorsement/

replace github.com/veraison/policy => ../../policy/

replace github.com/veraison/veraison/kvstore => ../../kvstore/

replace github.com/veraison/trustedservices => ../../trustedservices/

require (
	github.com/veraison/common v0.0.0
	github.com/veraison/endorsement v0.0.0
	github.com/veraison/policy v0.0.0
	github.com/veraison/trustedservices v0.0.0
	github.com/veraison/veraison/kvstore v0.0.0
	google.golang.org/grpc v1.41.0
)
