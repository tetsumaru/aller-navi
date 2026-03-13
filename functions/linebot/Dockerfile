FROM golang:1.26-bookworm AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o server ./cmd/main.go

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ghostscript \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/server /server

ENV PORT=8080
EXPOSE 8080

CMD ["/server"]
