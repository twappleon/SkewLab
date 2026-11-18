FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/skewbot ./cmd/bot \
 && CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/telegram-test ./cmd/telegram-test

FROM alpine:3.21
RUN apk add --no-cache su-exec \
 && addgroup -S app \
 && adduser -S -G app app \
 && mkdir -p /app/data \
 && chown app:app /app/data
WORKDIR /app
COPY --from=builder /out/skewbot /app/skewbot
COPY --from=builder /out/telegram-test /app/telegram-test
COPY docker/entrypoint.sh /app/entrypoint.sh
ENV PORT=8080
EXPOSE 8080
ENTRYPOINT ["/bin/sh", "/app/entrypoint.sh"]
