# syntax=docker/dockerfile:1
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o content-manager .

FROM alpine:3.18
WORKDIR /app
RUN apk add --no-cache curl
COPY --from=builder /app/content-manager .
CMD ["./content-manager"]
