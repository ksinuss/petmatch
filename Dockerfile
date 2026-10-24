# Сборка из корня репозитория (рядом с go.mod). SQLite — чистый Go (modernc), CGO не нужен.
FROM golang:1.25-bookworm AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY backend ./backend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /server ./backend/cmd/app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata wget \
	&& adduser -D -H -u 65532 app \
	&& mkdir -p /data && chown app:app /data

WORKDIR /data
COPY --from=builder /server /usr/local/bin/server
USER app

ENV DB_PATH=/data/data.db \
    UPLOAD_DIR=/data/uploads \
    GIN_MODE=release

EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s --start-period=10s CMD wget -qO- http://127.0.0.1:8080/health >/dev/null || exit 1
ENTRYPOINT ["/usr/local/bin/server"]
