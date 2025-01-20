# Start from the official Go image
FROM golang:alpine AS builder

# Add necessary security updates and certificates
RUN apk update && apk add --no-cache git

# Set the working directory
WORKDIR /app

# Copy files
COPY . .

# Build the app
RUN go get -d -v
RUN go build -o main .

# Expose port (note: this is just documentation)
EXPOSE 8080

# Command to run the application
CMD ["./main"]