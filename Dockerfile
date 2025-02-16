FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o palindrome ./cmd/main.go

FROM debian:bullseye-slim

WORKDIR /app

COPY --from=builder /app/palindrome /app/palindrome

CMD ["/app/palindrome"]
