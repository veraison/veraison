module github.com/veraison/cmd/endorsements

go 1.15

replace github.com/veraison/common => ../../common/

replace github.com/veraison/endorsement => ../../endorsement/

require (
	github.com/dolmen-go/flagx v0.0.0-20210127220802-bf12ea1664d9
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/stretchr/testify v1.6.1
	github.com/veraison/common v0.0.0
	github.com/veraison/endorsement v0.0.0
	go.uber.org/zap v1.10.0
)
