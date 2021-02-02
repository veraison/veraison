module veraison/dicetoken

go 1.15

replace veraison/common => ../../common/

replace veraison/dice => ../../../evidence/dice/

require (
	github.com/hashicorp/go-plugin v1.3.0
	github.com/stretchr/testify v1.6.1
	veraison/common v0.0.0
	veraison/dice v0.0.0
)
