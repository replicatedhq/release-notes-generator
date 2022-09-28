FROM golang:1.17-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o release-notes-generator

FROM alpine:latest

COPY --from=builder /app/release-notes-generator /release-notes-generator
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]