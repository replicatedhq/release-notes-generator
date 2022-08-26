FROM golang:1.17-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go build -o kots-release-helper

FROM alpine:latest

COPY --from=builder /app/kots-release-helper ./
COPY entrypoint.sh ./

ENTRYPOINT ["./entrypoint.sh"]