FROM golang:1.14.0 as golang_image

FROM golang_image as builder
COPY . /build
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o sensorpush_exporter main.go

FROM scratch
WORKDIR /
COPY --from=builder /build/sensorpush_exporter /
ENTRYPOINT ["/sensorpush_exporter"]
