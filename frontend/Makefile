include ../mk/common.mk

.PHONY: all
all: verifier

verifier:
	go build -o verifier cmd/veraison/main.go

.PHONY: clean
clean:
	rm -f verifier

.PHONY: run
run: build
	./verifier -p ../plugins/bin -d test/db
