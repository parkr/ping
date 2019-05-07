FROM golang as builder
WORKDIR /go/src/github.com/parkr/ping
EXPOSE 3306
COPY . .
RUN go version
RUN CGO_ENABLED=0 go install github.com/parkr/ping/...

FROM scratch
HEALTHCHECK --start-period=1ms --interval=30s --timeout=5s --retries=1 \
  CMD [ "/bin/ping-healthcheck" ]
COPY --from=0 /go/bin/ping-healthcheck /bin/ping-healthcheck
COPY --from=0 /go/bin/ping /bin/ping-server
ENTRYPOINT [ "/bin/ping-server" ]
