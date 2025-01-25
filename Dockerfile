# Use the official Go image
FROM golang:1.23

# Set the working directory
WORKDIR /app

# Initialize a new Go module directly in the container
# RUN go mod init app
COPY go.mod ./

# Install dependencies
RUN go get -u github.com/gin-gonic/gin \
    && go get github.com/golang-migrate/migrate/v4 \
    && go get github.com/lib/pq

# Copy the application source code into the container
COPY . .

# Download and verify dependencies
RUN go mod tidy

# Build the application
RUN go build -v -o main .

# Expose the application's port
EXPOSE 8080

# Command to run the application
CMD ["/app/main"]
