module github.com/veraison/frontend

go 1.15

replace github.com/veraison/common => ../common

replace github.com/veraison/policy => ../policy

replace github.com/veraison/trustedservices => ../trustedservices

replace github.com/veraison/verifier => ../verifier

replace github.com/veraison/veraison/kvstore => ../kvstore

require (
	github.com/gin-gonic/gin v1.7.0
	github.com/hashicorp/go-hclog v1.1.0
	github.com/mattn/go-sqlite3 v1.14.3
	github.com/moogar0880/problems v0.1.1
	github.com/veraison/common v0.0.0
	github.com/veraison/policy v0.0.0
	github.com/veraison/trustedservices v0.0.0
	github.com/veraison/veraison/kvstore v0.0.0
	github.com/veraison/verifier v0.0.0
	go.uber.org/zap v1.16.0
)
