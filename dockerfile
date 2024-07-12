# Use the official golang image from Docker Hub
FROM golang:1.22.3

# Set the working directory inside the container
WORKDIR /go/src/app/api-gateway

# Copy the Go application source code into the container
COPY ../api-gateway .

# Build the Go app (assuming the main package is in main.go)
RUN go build -o main .

# Expose port 4444 for the API gateway
EXPOSE 4444

# Command to run the executable
CMD ["./main"]

