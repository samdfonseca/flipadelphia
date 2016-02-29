default: build

build:
	go build

run:
	go build && ./flipadelphia

doc:
	godoc -http=:8888 -index

vet:
	go vet

