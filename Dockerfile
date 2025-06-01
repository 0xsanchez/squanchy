# Builder
FROM golang:latest AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o squanchy ./cmd

# Runner
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/squanchy .
CMD ["./squanchy"]