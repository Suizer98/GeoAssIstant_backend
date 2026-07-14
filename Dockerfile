FROM golang:1.25

RUN go install github.com/githubnemo/CompileDaemon@latest
RUN go install github.com/swaggo/swag/cmd/swag@latest

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download

COPY . /usr/src/app/

EXPOSE 10000

CMD ["CompileDaemon", "--build=go build -buildvcs=false -o main .", "--command=./main", "--polling"]
