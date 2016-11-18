default: build

deps:
	godep save ./...

build:
	go build
	
run:
	go build && ./flipadelphia

doc:
	godoc -http=:8888 -index

vet:
	go vet

dev:
	go build && ./flipadelphia serve --env development

noauth:
	go build && ./flipadelphia serve --env noauth

check:
	go build && ./flipadelphia sanitycheck

install:
	mv flipadelphia /usr/bin/

test:
	- go test -v ./config
	- go test -v ./store
	- go test -v ./server
	- go test -v ./cmd/flippy

debug:
	go build && dlv exec ./flipadelphia s
