module veraison/opapolicyengine

go 1.15

replace veraison/common => ../../common/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/open-policy-agent/opa v0.22.0
	github.com/stretchr/testify v1.6.1
	veraison/common v0.0.0
)
