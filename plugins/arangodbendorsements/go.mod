module arangodbendorsements

go 1.15

replace github.com/veraison/common => ../../common/

require (
	github.com/arangodb/go-driver v0.0.0-20200831144536-17278d36b7e8
	github.com/golang/mock v1.4.4
	github.com/hashicorp/go-plugin v1.3.0
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0-00010101000000-000000000000
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
