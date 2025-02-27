FROM golang:1.20 AS builder

WORKDIR /app

COPY . .

RUN go mod tidy

RUN go build -o db_backup

FROM alpine:latest

RUN apk add --no-cache mysql-client postgresql-client

WORKDIR /root/

COPY --from=builder /app/db_backup .

CMD ["./db_backup"]
