module github.com/veraison/sqliteendrosements

go 1.15

replace github.com/veraison/policy => ../../policy/

replace github.com/veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.6.1
	github.com/veraison/common v0.0.0
	github.com/veraison/policy v0.0.0
)
