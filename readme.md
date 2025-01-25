# GeoAssIstant Backend API

## Local Development

### Spin up service
In terminal, type command below to perform different actions:

```
docker-compose up --build
docker-compose up
docker-compose down -v
```

### Update Go modules

I don't have any Golang install on my local machine so I did this to obtain `go.sum` and `go.mod`:

```
docker build -t geoassistant_api . && \
docker create --name temp_geoassistant_api geoassistant_api && \
docker cp temp_geoassistant_api:/usr/src/app/go.mod ./go.mod && \
docker cp temp_geoassistant_api:/usr/src/app/go.sum ./go.sum && \
docker cp temp_geoassistant_api:/usr/src/app/docs/. ./docs && \
docker rm temp_geoassistant_api
```

### Update Swagger docs

```
docker exec geoassistant_api swag init
```

### Check linting and typing

External pull to check the codes:

```
docker run --rm -v $(pwd):/app -w /app golangci/golangci-lint:v1.53.3 golangci-lint run
```

To fix with `gofmt` issues:

```
docker run --rm -v $(pwd):/app -w /app golang:1.23 gofmt -s -w .
```