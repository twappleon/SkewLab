FROM golang:1.23-alpine AS builder
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/skewbot ./cmd/bot
FROM alpine:3.21
RUN addgroup -S app && adduser -S -G app app && mkdir -p /app/data && chown app:app /app/data
WORKDIR /app
COPY --from=builder /out/skewbot /app/skewbot
USER app
ENV PORT=8080
EXPOSE 8080
CMD ["/app/skewbot","--data-dir","/app/data"]
