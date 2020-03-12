FROM golang:1.14.0 as tester
COPY . /build
WORKDIR /build
RUN go test

FROM golang:1.14.0 as builder
COPY . /build
WORKDIR /build
RUN go build -o opcua_exporter main.go

FROM alpine:latest
WORKDIR /root
COPY --from=builder /build/opcua_exporter .
COPY --from=builder /build/nodes.json .
ENTRYPOINT ["./opcua_exporter"]
CMD ["-config", "nodes.json"]
