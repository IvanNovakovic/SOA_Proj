# Multi-stage Dockerfile for follower-service
# Builds a Go binary in a builder image then copies it into a minimal runtime image
# This version intentionally does not install `git` in the builder; if your modules
# require git access during `go mod download`, switch to a builder image that includes
# git or add git in the builder stage.

FROM golang:1.20-alpine AS builder
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

# Copy source and build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags='-s -w' -o /out/follower-service ./

########## Runtime image ##########
FROM alpine:3.18 AS runtime
RUN apk add --no-cache ca-certificates

COPY --from=builder /out/follower-service /usr/local/bin/follower-service

EXPOSE 8082

ENTRYPOINT ["/usr/local/bin/follower-service"]
