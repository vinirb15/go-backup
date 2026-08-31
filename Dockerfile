FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o db_backup

FROM alpine:latest

# postgresql-client -> pg_dump / pg_restore ; mariadb-client -> mysqldump / mysql
# tzdata is required for the TZ env var (time.LoadLocation) to work.
RUN apk add --no-cache postgresql-client mariadb-client tzdata

WORKDIR /root/

COPY --from=builder /app/db_backup .

CMD ["./db_backup"]
