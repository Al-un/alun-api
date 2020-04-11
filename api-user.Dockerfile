# https://hub.docker.com/layers/golang/library/golang/1.13.10-alpine/images/sha256-cddae2b986dfd92581a082beeb4e7898c8eaa4ac93b618d2dacc5c14983abc20?context=explore
FROM golang:1.13.10-alpine AS builder

WORKDIR /usr/src/app

# No port exposition because it is defined by environment variable

# Dependencies
COPY ./go.mod .
COPY ./go.sum .
# https://stackoverflow.com/a/56693289/4906586
RUN go mod download

# Copy sources and build the binary
COPY . .
RUN go build -o api-user cmd/alun-user/main.go

# https://docs.docker.com/develop/develop-images/multistage-build/#use-multi-stage-buildssudo
FROM alpine:latest AS runner

WORKDIR /usr/local/bin/

# Copy binary
COPY --from=builder /usr/src/app/api-user .
# Copy email templates
COPY ./alun/utils/email_templates/user_* ./alun/utils/email_templates/

CMD ["./api-user"]