# https://hub.docker.com/_/golang?tab=tags
FROM golang:1.14-alpine AS builder

WORKDIR /usr/src/app

# No port exposition because it is defined by environment variable

# Dependencies
COPY ./go.mod .
COPY ./go.sum .
# https://stackoverflow.com/a/56693289/4906586
RUN go mod download

# Copy sources and build the binary
COPY . .
CMD ["go", "run", "cmd/alun-api/main.go"]