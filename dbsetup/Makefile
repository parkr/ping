all: deps fmt build

fmt:
	go fmt ./...

build: deps
	go build .

deps:
	go get github.com/go-sql-driver/mysql \
		github.com/jmoiron/sqlx \
		github.com/parkr/gossip/serializer
