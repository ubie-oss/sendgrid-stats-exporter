FROM golang:1.17.6-bullseye as builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o exporter

FROM debian:bullseye-slim
RUN set -x && apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
        ca-certificates && \
        rm -rf /var/lib/apt/lists/*
COPY --from=builder /app/exporter /app/exporter
CMD ["/app/exporter"]
