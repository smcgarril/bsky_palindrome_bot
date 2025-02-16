ARG GO_VERSION=1
FROM golang:${GO_VERSION}-bookworm as builder
RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o palindrome ./cmd/main.go

FROM golang:1.23

COPY --from=builder /etc/ssl/certs /etc/ssl/certs
COPY --from=builder /app/palindrome /app/palindrome

EXPOSE 8080

WORKDIR /app

CMD ["/app/palindrome"]
