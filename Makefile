all: deps fmt build test

fmt:
	go fmt ./...

build: deps
	go build .

testdeps:
	script/setup-test-database

test: deps testdeps
	go test ./...

deps:
	go get github.com/go-sql-driver/mysql \
		github.com/jmoiron/sqlx \
		github.com/parkr/gossip/serializer
