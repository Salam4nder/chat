ARG GO_VERSION=1.20
FROM golang:${GO_VERSION} AS build

WORKDIR /app

COPY . .
# COPY internal/db/migrations /app/db/migrations

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main ./cmd/chat/

FROM scratch

WORKDIR /app

COPY --from=build /app/main /app/main
# COPY --from=build /app/db/migrations /app/db/migrations
# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

CMD ["./main"]
