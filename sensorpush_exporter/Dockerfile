FROM alpine:latest as certs
RUN apk --update add ca-certificates

FROM golang:1.14.0 as builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sensorpush_exporter main.go

FROM scratch
ENV PATH=/bin
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /build/sensorpush_exporter /bin/
ENTRYPOINT ["/bin/sensorpush_exporter"]
