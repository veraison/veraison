module github.com/veraison/cmd/policy

go 1.15

replace github.com/veraison/common => ../../common/

replace github.com/veraison/policy => ../../policy/

require (
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	github.com/veraison/policy v0.0.0
	go.uber.org/zap v1.17.0
)
