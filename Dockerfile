FROM golang:1.24.1 as builder

# basic setup
WORKDIR /app

# get dependencies
COPY go.mod go.sum ./
RUN go mod download

# build app
COPY . ./
RUN go build -trimpath ./cmd/at-mirror

# get cert files
FROM alpine:latest as certs
RUN apk --update add ca-certificates

# build runtime image
FROM debian:stable-slim
RUN apt update && apt install -y curl
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /app/at-mirror .
ENTRYPOINT ["./at-mirror"]
