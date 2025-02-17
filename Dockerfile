FROM golang:1.23

RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /usr/src/app
COPY . /usr/src/app/

# Initialize a Go module inside the container
RUN go mod init geoai-app || true

# Go Gin dependencies
RUN go mod init geoai-app || true && \
    go mod tidy && \
    go get github.com/gin-gonic/gin \
           github.com/joho/godotenv \
           github.com/golang-migrate/migrate/v4 \
           github.com/golang-migrate/migrate/v4/database/postgres \
           github.com/golang-migrate/migrate/v4/source/file \
           github.com/lib/pq \
           github.com/swaggo/swag \
           github.com/swaggo/gin-swagger \
           github.com/swaggo/files

EXPOSE 8080

# Start the application with CompileDaemon for hot reload
CMD ["CompileDaemon", "--build=go build -o main .", "--command=./main", "--polling"]
