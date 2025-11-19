FROM golang:1.25.4-alpine3.22 AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o /app/bin/app ./cmd/avito/main.go

RUN go install github.com/rubenv/sql-migrate/...@latest


FROM alpine:latest

WORKDIR /app

RUN apk add --no-cache bash wget git

RUN wget https://raw.githubusercontent.com/vishnubob/wait-for-it/master/wait-for-it.sh && \
    chmod +x wait-for-it.sh

COPY --from=builder /app/bin/app /app/bin/app
COPY migrations/ /app/migrations/
COPY migrate.yaml /app/migrate.yaml
COPY entrypoint.sh /app/entrypoint.sh

RUN chmod +x /app/entrypoint.sh

COPY --from=builder /go/bin/sql-migrate /usr/local/bin/sql-migrate

ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["/app/bin/app"]