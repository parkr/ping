FROM golang:1.24.5-bookworm as builder
WORKDIR /workspace
EXPOSE 3306
COPY . .
RUN go version
RUN go install github.com/parkr/ping/... && ls -l /go/bin/ping && ls -l /go/bin/ping-healthcheck

FROM debian:bookworm-slim
HEALTHCHECK --start-period=1s --interval=30s --timeout=5s --retries=1 \
  CMD [ "/go/bin/ping-healthcheck" ]
COPY --from=builder /go/bin/* /go/bin/
ENTRYPOINT [ "/go/bin/ping" ]
