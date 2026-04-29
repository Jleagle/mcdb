# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy go.mod and go.sum, then download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application
# -ldflags="-w -s" strips debug information for a smaller binary
# CGO_ENABLED=0 and GOOS=linux ensure a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o mcdb ./main.go

# Stage 2: Create the final lean image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/mcdb /app/mcdb

# Copy the web templates, which are required at runtime
COPY --from=builder /app/web/templates ./web/templates

# Expose the port the web server listens on
EXPOSE 8080

# Set environment variable for MongoDB connection (runtime configuration)
# It's recommended to pass this during 'docker run' or via docker-compose
ENV MCDB_MONGO=""

# Define the command to run the web server
CMD ["./mcdb", "serve"]
