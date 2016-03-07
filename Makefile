default: build

deps:
	godep save

build:
	go build
	
run:
	go build && ./flipadelphia serve

doc:
	godoc -http=:8888 -index

vet:
	go vet

dev:
	go build && ./flipadelphia serve --env development

check:
	go build && ./flipadelphia sanitycheck

install:
	mv flipadelphia /usr/bin/

test:
	go test -v ./config && go test -v ./store
