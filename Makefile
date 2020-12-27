PKG=github.com/parkr/ping
REV:=$(shell git rev-parse HEAD)

all: fmt build test

fmt:
	go fmt $(PKG)/...

build:
	go build $(PKG)

testdeps:
	go build $(PKG)/cmd/ping-initialize-db

test: testdeps
	. script/test-env && go test -v ./...

docker-build:
	docker build -t parkr/ping:$(REV) .

docker-test: docker-build
	docker run --name=ping-test --rm -it --net=host parkr/ping:$(REV)

docker-release: docker-build
	docker push parkr/ping:$(REV)

dive: docker-build
	dive parkr/ping:$(REV)
