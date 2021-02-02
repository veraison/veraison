module veraison/verifier

go 1.15

replace veraison/common => ../common/

replace veraison/endorsement => ../endorsement/

replace veraison/policy => ../policy/

require (
	github.com/hashicorp/go-hclog v0.0.0-20180709165350-ff2cf002a8dd
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/stretchr/testify v1.6.1
	golang.org/x/tools/gopls v0.5.0 // indirect
	veraison/common v0.0.0
	veraison/endorsement v0.0.0
	veraison/policy v0.0.0
)
