module arangodbendorsements

go 1.15

replace veraison/common => ../../common/

require (
	github.com/arangodb/go-driver v0.0.0-20200831144536-17278d36b7e8
	github.com/hashicorp/go-plugin v1.3.0
	github.com/mattn/go-sqlite3 v1.14.4
	github.com/stretchr/testify v1.6.1
	veraison/common v0.0.0-00010101000000-000000000000
)
