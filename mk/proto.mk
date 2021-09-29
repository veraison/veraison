# Copyright 2021 Contributors to the Veraison project.
# SPDX-License-Identifier: Apache-2.0
#
# variables:
# * PROTOSRCS  - a list of .proto source files to compile
# * PROTOPATHS - a list of paths, in addition to the current directory,  that
#                will be searched for the .proto source files and their
#                includes.
# targets:
# * all        - Build .pb.go file from .proto [DEFAULT]
#
# note: This should be includded before test.mk and/or plugin.mk
#

CLEANFILES += protogen *.pb.go *.pb.json.go
GEN_FILES += protogen

all: protogen

protogen: $(PROTOSRCS)
	protoc --go_out=. --go_opt=paths=source_relative --proto_path=. \
		$(foreach path,$(PROTOPATHS), --proto_path=$(path)) \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--go-json_out=. --go-json_opt=paths=source_relative $(PROTOSRCS)
	@echo ".pb.go generated" > protogen
