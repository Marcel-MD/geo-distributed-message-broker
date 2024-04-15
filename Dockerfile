# Start from golang base image
FROM golang:1.21.5 as builder

# Add Maintainer info
LABEL maintainer="Marcel Vlasenco"

# Set the current working directory inside the container 
WORKDIR /app

# Copy go mod and sum files 
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and the go.sum files are not changed 
RUN go mod download 

# Copy the source from the current directory to the working Directory inside the container 
COPY . .

# Build the Go app
RUN CGO_ENABLED=1 GOOS=linux go build -o main -a -ldflags '-linkmode external -extldflags "-static"' .

# Start a new stage from scratch
FROM alpine:3.19.0
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main ./

# Command to run the executable
CMD ["./main"]