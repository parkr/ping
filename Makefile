all: fmt build test

fmt:
	go fmt ./...

build:
	go build .

testdeps:
	go build ./cmd/ping-initialize-db
	script/setup-test-database

test: testdeps
	. script/test-env && go test ./...

docker-release:
	docker build -t parkr/ping:$(shell git rev-parse HEAD) .
	docker push parkr/ping:$(shell git rev-parse HEAD)
