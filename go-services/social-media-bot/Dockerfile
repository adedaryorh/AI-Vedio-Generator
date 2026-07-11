# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o social-media-bot main.go

FROM alpine:3.18
WORKDIR /app
COPY --from=builder /app/social-media-bot .
COPY --from=builder /app/shared ./shared
CMD ["./social-media-bot"] 