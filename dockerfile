# Use the official golang image from Docker Hub
FROM golang:1.22.3

# Set the current working directory inside the container
WORKDIR /go/src/app

# Copy the 'api-gateway' directory into the container
COPY api-gateway /go/src/app/api-gateway

# Build the Go app
RUN go build -o main .

# Expose port 4444 for API gateway
EXPOSE 4444

# Command to run the executable
CMD ["./main"]

