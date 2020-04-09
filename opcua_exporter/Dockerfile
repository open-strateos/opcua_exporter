FROM golang:1.14.0 as golang_image

FROM golang_image as tester
COPY . /build
WORKDIR /build
RUN go test

FROM golang_image as builder
COPY --from=tester /build /build
COPY --from=tester /go /go
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o opcua_exporter main.go opc_value_handler.go opcua_bitvector_handler.go

FROM scratch
WORKDIR /
COPY --from=builder /build/opcua_exporter /
COPY --from=builder /build/nodes.json /
ENTRYPOINT ["/opcua_exporter"]
CMD ["-config", "nodes.json"]
