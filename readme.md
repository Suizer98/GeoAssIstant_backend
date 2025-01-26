# Geo AssIstant Backend API

## Description

Geo Assistant is a conversational system designed to provide geospatial assistance and plans. It integrates a Groq API, allowing users to engage with different AI models to solve geospatial queries efficiently. The system supports maintaining conversation history and user management.

Staging development side: https://geoassistant-backend.onrender.com/swagger/index.html

Currently it uses both free tier web service and postgressql instance from Render, and limited requests are allowed for GroqCloud API.

## Tech stacks:
![Tech Stacks](https://skillicons.dev/icons?i=go,postgres,docker,bash)

## Staging Host/Development
To deploy this on `Render` dashboard, simply create web service and set language to `Dockerfile`

## Local Development

### Spin up service
In terminal, type command below to perform different actions:

```sh
docker-compose up --build # Build and run service
docker-compose up # Run service with caches
docker-compose down -v # Clamp down containers
```

The local service can be found in https://localhost:8080/swagger

### Update Go modules

I don't have any Golang install on my local machine so I did this to obtain `go.sum` and `go.mod`:

```sh
docker build -t geoassistant_api . && \
docker create --name temp_geoassistant_api geoassistant_api && \
docker cp temp_geoassistant_api:/usr/src/app/go.mod ./go.mod && \
docker cp temp_geoassistant_api:/usr/src/app/go.sum ./go.sum && \
docker cp temp_geoassistant_api:/usr/src/app/docs/. ./docs && \
docker rm temp_geoassistant_api
```

### Update Swagger docs

```sh
docker exec geoassistant_api swag init -g app/app.go
```

### Check linting and typing

External pull to check the codes:

```sh
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run
```

To fix with `gofmt` issues:

```sh
docker run --rm -v $(pwd):/app -w /app golang:1.23 gofmt -s -w .
```