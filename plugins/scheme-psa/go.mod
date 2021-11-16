module veraison/veraison/scheme/psa

go 1.15

replace github.com/veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/stretchr/testify v1.7.0
	github.com/veraison/common v0.0.0
	github.com/veraison/psatoken v0.0.1
	google.golang.org/protobuf v1.27.1
)
