# Use the official Golang image as a base
FROM golang:1.21.0

# Set the working directory for subsequent instructions
WORKDIR /app

# Copy the Go modules definition files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source code into the container
# Note: Adjust this if you have a more complex directory structure
COPY . .

# Build the application, specifying the main file
# Adjust the output binary name if necessary
RUN CGO_ENABLED=0 GOOS=linux go build -o /docker-gs-ping ./cmd/main.go

# Expose the port that the application will listen on
EXPOSE 8080

# Specify the command to run the executable
CMD ["/docker-gs-ping"]
