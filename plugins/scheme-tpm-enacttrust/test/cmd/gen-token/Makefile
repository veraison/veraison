.DEFAULT_GOAL := gen-token

CLEANFILES += gen-token

gen-token: main.go
	go build -o gen-token

test: ;
.PHONY: test

include ../../../../../mk/common.mk
include ../../../../../mk/lint.mk
include ../../../../../mk/pkg.mk
