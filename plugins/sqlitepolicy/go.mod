module github.com/veraison/sqlitepolicy

go 1.15

replace github.com/veraison/policy => ../../policy/

replace github.com/veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	golang.org/x/sys v0.0.0-20210309074719-68d13333faf2 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
