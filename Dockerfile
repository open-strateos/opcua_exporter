FROM golang:1.14.0-alpine3.11 as builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o opcua_exporter main.go

FROM scratch
WORKDIR /
COPY --from=builder /build/opcua_exporter /
COPY --from=builder /build/nodes.json /
ENTRYPOINT ["/opcua_exporter"]
CMD ["-config", "nodes.json"]
