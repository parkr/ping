FROM golang
WORKDIR /go/src/github.com/parkr/ping
EXPOSE 3306
COPY . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/ping/...

FROM scratch
COPY --from=0 /go/bin/ping /bin/ping-server
ENTRYPOINT [ "/bin/ping-server" ]
