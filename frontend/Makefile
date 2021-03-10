include ../mk/common.mk
include ../mk/lint.mk
include ../mk/test.mk

.PHONY: all
all: verifier

verifier: cmd/veraison/main.go
	go build -o $@ $<

.PHONY: clean
clean:
	rm -f verifier

.PHONY: run
run: verifier
	./verifier -p ../plugins/bin -d test/db
