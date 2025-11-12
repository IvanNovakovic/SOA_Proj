# build stage
FROM golang:1.25-alpine AS builder
WORKDIR /app

# copy go mod and sum files
COPY src/go.mod src/go.sum ./src/
WORKDIR /app/src
RUN go mod download

# copy the project and build the binary from src/
COPY src/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/stakeholders-service main.go

# runtime stage
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/stakeholders-service ./stakeholders-service

# ENV PORT=8080
EXPOSE 8080

CMD ["./stakeholders-service"]