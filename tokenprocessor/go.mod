module github.com/veraison/tokenprocessor

go 1.15

replace github.com/veraison/common => ../common

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.5
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
