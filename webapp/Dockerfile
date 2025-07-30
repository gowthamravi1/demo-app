# Build stage
FROM golang:1.20-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o webapp main.go

# Final stage
FROM alpine:3.17
RUN apk add --no-cache ca-certificates
WORKDIR /root/
COPY --from=builder /app/webapp .
CMD ["./webapp"]
