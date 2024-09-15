# Official Go image
FROM golang:1.23.1-alpine

# Install bash for the wait-for-it script
RUN apk add --no-cache bash

# set the working directory
WORKDIR /app

# Copy the dependencies file to the working directory
COPY go.mod go.sum ./

# Download the dependencies
RUN go mod download

# Copy the source code into the container  
COPY . .

# Build the application
RUN go build -o main ./cmd/api

# Add wait-for-it script for DB readiness check
COPY wait-for-it.sh /wait-for-it.sh
RUN chmod +x /wait-for-it.sh

# Expose the port the app runs on
EXPOSE 8080

# Run the application
CMD ["./main"]
