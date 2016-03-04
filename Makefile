default: build

deps:
	glide install

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
