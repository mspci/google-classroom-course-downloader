# Use the official Golang image as the base image
FROM golang:1.20

# Set the working directory
WORKDIR /app

# Copy the source files
COPY . .

# Build the application
RUN go build -o main .

# Expose the port that the application uses
EXPOSE 8080

# Command to run the executable
CMD ["./main"]
