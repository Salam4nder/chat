ARG GO_VERSION=1.20
FROM golang:${GO_VERSION} AS build

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o main ./cmd/chat/

FROM scratch

WORKDIR /app

COPY --from=build /app/main /app/main
COPY --from=build /app/config.yaml /app/config.yaml
COPY --from=build /app/internal/db/cql /app/internal/db/cql
# COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

CMD ["./main"]
