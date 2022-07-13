FROM golang:1.18.4-bullseye as builder
WORKDIR /go/src/github.com/parkr/ping
EXPOSE 3306
COPY . .
RUN go version
RUN go install github.com/parkr/ping/... && ls -l /go/bin

FROM debian:bullseye-slim
HEALTHCHECK --start-period=1s --interval=30s --timeout=5s --retries=1 \
  CMD [ "/go/bin/ping-healthcheck" ]
COPY --from=builder /go/bin/* /go/bin/
ENTRYPOINT [ "/go/bin/ping" ]
