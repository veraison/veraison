module github.com/verason/policymanager

go 1.15

replace github.com/veraison/common => ../common

require (
	github.com/hashicorp/go-hclog v0.0.0-20180709165350-ff2cf002a8dd
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/ohler55/ojg v1.2.1 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/veraison/common v0.0.0
	google.golang.org/grpc v1.27.1 // indirect
)
