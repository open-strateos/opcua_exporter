FROM golang:1.14.0 as golang_image

FROM golang_image as tester
COPY . /build
WORKDIR /build
RUN go test

FROM golang_image as builder
COPY --from=tester /build /build
COPY --from=tester /go /go
WORKDIR /build
RUN go build -o opcua_exporter main.go

FROM alpine:latest
WORKDIR /root
COPY --from=builder /build/opcua_exporter .
COPY --from=builder /build/nodes.json .
ENTRYPOINT ["./opcua_exporter"]
CMD ["-config", "nodes.json"]
