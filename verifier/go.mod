module github.com/veraison/verifier

go 1.15

replace github.com/veraison/common => ../common/

replace github.com/veraison/endorsement => ../endorsement/

replace github.com/veraison/policy => ../policy/

require (
	github.com/hashicorp/go-hclog v0.0.0-20180709165350-ff2cf002a8dd
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	github.com/veraison/endorsement v0.0.0
	github.com/veraison/policy v0.0.0
	go.uber.org/zap v1.16.0
	golang.org/x/tools/gopls v0.5.0 // indirect
)
