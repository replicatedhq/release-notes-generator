FROM golang:1.17-alpine as builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN go build -o release-notes-generator

FROM alpine:latest

COPY --from=builder /app/release-notes-generator ./
COPY entrypoint.sh ./

ENTRYPOINT ["./entrypoint.sh"]