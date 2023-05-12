# Build stage
FROM golang:1.17-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum main.go ./

# Install dependencies
RUN apk add --no-cache git build-base 
RUN go mod download

# Copy the source code
COPY . .

# Build the binary for ARM64 architecture
RUN GOARCH=arm64 GOOS=linux CGO_ENABLED=0 go build -o billing main.go

# Runtime stage
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the binary from the builder image
COPY --from=builder /app/billing .

# Expose port 8080
EXPOSE 8080

# Run the binary and mount the index.html file
CMD ["/app/billing"]
