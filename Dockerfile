# Étape 1 : build
FROM golang:1.23 AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# Compilation statique (pour éviter les problèmes liés à GLIBC dans Debian slim)
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o forum ./cmd

# Étape 2 : image minimale
FROM debian:bullseye-slim

WORKDIR /app
COPY --from=builder /app/forum .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

EXPOSE 8080
CMD ["./forum"]
