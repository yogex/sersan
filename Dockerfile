# build stage
FROM golang:1.13-alpine AS build-env
RUN mkdir -p /go/src/github.com/salestock/sersan
WORKDIR /go/src/github.com/salestock/sersan
COPY . .
RUN go build -o server

# final stage
FROM alpine
WORKDIR /app
RUN mkdir -p config
COPY ./config/grids.yaml ./config/grids.yaml
COPY  --from=build-env /go/src/github.com/salestock/sersan/server /app/
ENTRYPOINT ./server
