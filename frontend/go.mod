module github.com/veraison/frontend

go 1.15

replace github.com/veraison/common => ../common

replace github.com/veraison/endorsement => ../endorsement

replace github.com/veraison/policy => ../policy

replace github.com/veraison/tokenprocessor => ../tokenprocessor

replace github.com/veraison/verifier => ../verifier

require (
	github.com/gin-gonic/gin v1.7.0
	github.com/moogar0880/problems v0.1.1
	github.com/veraison/common v0.0.0
	github.com/veraison/tokenprocessor v0.0.0-00010101000000-000000000000
	github.com/veraison/verifier v0.0.0
	go.uber.org/zap v1.16.0
)
