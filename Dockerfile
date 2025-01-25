FROM golang:1.23

RUN go install github.com/githubnemo/CompileDaemon@latest

WORKDIR /usr/src/app
COPY . /usr/src/app/

# Initialize a Go module inside the container
RUN go mod init geoai-app || true
RUN go get github.com/gin-gonic/gin \
    && go get github.com/joho/godotenv \
    && go get github.com/golang-migrate/migrate/v4 \
    && go get github.com/golang-migrate/migrate/v4/database/postgres \
    && go get github.com/golang-migrate/migrate/v4/source/file \
    && go get github.com/lib/pq \
    && go mod tidy

EXPOSE 8080

# Start the application with CompileDaemon for hot reload
CMD ["CompileDaemon", "--build=go build -o main .", "--command=./main", "--polling"]
