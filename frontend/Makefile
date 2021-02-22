
build:
	go build cmd/veraison/main.go

clean:
	rm -f ./main

run: build
	./main -p ../plugins/bin -d test/db
