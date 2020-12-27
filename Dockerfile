FROM golang:1.14.3 as builder
WORKDIR /go/src/github.com/parkr/ping
EXPOSE 3306
COPY . .
RUN go version
RUN go install github.com/parkr/ping/...

FROM scratch
HEALTHCHECK --start-period=1s --interval=30s --timeout=5s --retries=1 \
  CMD [ "/bin/ping-healthcheck" ]
COPY --from=builder /go/bin/ping-healthcheck /bin/ping-healthcheck
COPY --from=builder /go/bin/ping /bin/ping-server
ENTRYPOINT [ "/bin/ping-server" ]
