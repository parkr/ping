FROM golang as builder
WORKDIR /go/src/github.com/parkr/ping
EXPOSE 3306
COPY . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/ping/...

FROM golang as curler
RUN CGO_ENABLED=0 go get github.com/parkr/go-curl/cmd/go-curl

FROM scratch
HEALTHCHECK --interval=30s --timeout=3s \
  CMD [ "/bin/go-curl", "-f", "http://127.0.0.1:8000/_health" ]
COPY --from=curler /go/bin/go-curl /bin/go-curl
COPY --from=builder /go/bin/ping /bin/ping-server
ENTRYPOINT [ "/bin/ping-server" ]
