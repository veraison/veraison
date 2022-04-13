module github.com/veraison/trustedservices

go 1.15

replace github.com/veraison/common => ../common/

replace github.com/veraison/policy => ../policy/

replace github.com/veraison/veraison/kvstore => ../kvstore/

require (
	github.com/veraison/common v0.0.0
	github.com/veraison/veraison/kvstore v0.0.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)
