# Étape 1 : build
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o forum ./cmd

# Étape 2 : image minimale
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/forum .

EXPOSE 8080
CMD ["./forum"]