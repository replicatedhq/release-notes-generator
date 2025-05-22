FROM cgr.dev/chainguard/go:latest AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
RUN go build -o release-notes-generator

FROM cgr.dev/chainguard/wolfi-base:latest

COPY --from=builder /app/release-notes-generator /release-notes-generator
COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]