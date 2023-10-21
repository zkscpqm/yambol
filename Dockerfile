
# Step 1: Build the Go application in a full Go environment
FROM golang:1.21 as builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o yambol ./cmd/main.go

# Step 2: Use a minimal image for running the application
FROM alpine:latest

WORKDIR /app/

COPY --from=builder /app/config.json .
COPY --from=builder /app/yambol .

# HTTP
EXPOSE 21419

# HTTPS. Nice
EXPOSE 21420

# gRPC
EXPOSE 21421

# gRPC +TLS
EXPOSE 21422

ENTRYPOINT ["./yambol"]
