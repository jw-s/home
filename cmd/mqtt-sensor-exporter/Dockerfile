FROM golang:1.14.0-stretch as builder

RUN adduser --disabled-password -u 10001 exporter

WORKDIR /go/src/github.com/jw-s/home

RUN apt-get update && \
    apt-get install -y make git ca-certificates

COPY . .

RUN CGO_ENABLED=0 GOARCH=arm go build -trimpath -a -ldflags="-w -s" -o ./bin/sensor-exporter ./cmd/mqtt-sensor-exporter

FROM scratch

WORKDIR /
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /go/src/github.com/jw-s/home/bin/sensor-exporter /sensor-exporter
USER exporter
EXPOSE 8080
ENTRYPOINT ["/sensor-exporter"]