FROM golang:1.20-alpine AS builder

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o forum ./cmd

FROM alpine:latest

WORKDIR /app
COPY --from=builder /app/forum .

CMD ["./forum"]
