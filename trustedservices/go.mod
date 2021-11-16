module github.com/veraison/trustedservices

go 1.15

replace github.com/veraison/common => ../common/

replace github.com/veraison/policy => ../policy/

replace github.com/veraison/veraison/kvstore => ../kvstore/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	github.com/veraison/policy v0.0.0
	github.com/veraison/veraison/kvstore v0.0.0
	go.uber.org/zap v1.16.0
	golang.org/x/tools v0.0.0-20200914163123-ea50a3c84940 // indirect
	google.golang.org/grpc v1.41.0
	google.golang.org/protobuf v1.27.1
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	honnef.co/go/tools v0.0.1-2020.1.5 // indirect
)
