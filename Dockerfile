# Use a minimal base image
FROM golang:1.20-alpine AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod ./

# Download the Go module dependencies
RUN go mod download && go mod verify

# Copy the source code
COPY . ./

# Build the Go binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# Use a minimal base image for runtime
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/app ./

# Set the executable permissions for the binary
RUN chmod +x ./app

# Expose the port that the web service listens on
EXPOSE 80

# Set the entrypoint command
CMD ["./app"]
