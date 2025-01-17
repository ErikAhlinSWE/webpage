# Start from the official Go image
FROM golang:alpine AS builder

# Add necessary security updates and certificates
RUN apk update && apk add --no-cache git
WORKDIR $GOPATH/src/mypackage/myapp/
COPY . .

COPY . .
RUN go get -d -v
RUN go build -o /app/cmd/main
# Set the working directory





# Expose port (note: this is just documentation)
EXPOSE 8080

# Command to run the application
CMD ["./main"]