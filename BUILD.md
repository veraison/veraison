# Make targets

The following `make(1)` targets can be invoked from everywhere:

* `make all` build all the plugins
* `make clean` remove all plugins
* `make test` run the unit tests with coverage information
* `make lint` run the code linter (requires [golangci-lint](https://golangci-lint.run/usage/install/)
* `make lint-extra` run the code linter with extra checks

# gRPC Tooling

* [follow the instructions](https://grpc.io/docs/protoc-installation/) to install the protocol buffer compiler for your build platform
* install the plugins:
```
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
go install github.com/mitchellh/protoc-gen-go-json@latest
```
