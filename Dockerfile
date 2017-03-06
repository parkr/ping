FROM golang

WORKDIR /go/src/github.com/parkr/ping

EXPOSE 3306

ADD . .

RUN go version

# Compile a standalone executable
RUN CGO_ENABLED=0 go install github.com/parkr/ping/...

# Run the ping command by default.
CMD [ "ping" ]
