FROM golang:1.14.0-alpine3.11 as builder
COPY . /build
WORKDIR /build
RUN go build -o opcua_exporter main.go

FROM alpine:latest
WORKDIR /root
COPY --from=builder /build/opcua_exporter .
ENTRYPOINT ["./opcua_exporter"]
CMD []