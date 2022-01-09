OUTPUT := ../test_vectors.go

DEPS := $(wildcard Comid*.json)

all: $(OUTPUT)

$(OUTPUT): $(DEPS)
	env TV_DOT_GO=$(OUTPUT) ./build-test-vectors.sh

clean: ; $(RM) -f *.cbor
