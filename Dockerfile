# Stage 1: Build the Go binary
FROM golang:1.20-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files to the workspace
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY ./cmd ./cmd

# Build the Go app
RUN go build -o /app/main ./cmd/main.go

# Stage 2: Create a small image with the binary
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
