# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk --no-cache add ca-certificates git

# Set the working directory inside the container
WORKDIR /app

# Copy the Go module files and download dependencies
COPY go.mod go.sum ./

# Install dependencies
RUN go mod download

# Copy the rest of the application source code
COPY . .

# Build the Go application with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a \
    -installsuffix cgo \
    -ldflags="-w -s" \
    -o /app/server \
    ./cmd/main.go

# Stage 2: Create a minimal image to run the application
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates wget

# Create non-root user and group first
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set the working directory
WORKDIR /app

# Copy the built Go application from the builder stage
COPY --from=builder /app/server .

# Change ownership to non-root user
RUN chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Command to run the application
CMD ["./server"]