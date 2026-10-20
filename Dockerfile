# syntax=docker/dockerfile:1

# The builder runs on the build machine's platform and cross-compiles for the
# target (pure Go, CGO off), so multi-arch builds never compile under QEMU.
FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/db_backup .

FROM alpine:3.24

# postgresql-client -> pg_dump / pg_restore ; mariadb-client -> mysqldump / mysql
# tzdata is required for the TZ env var (time.LoadLocation) to work.
RUN apk add --no-cache postgresql-client mariadb-client tzdata \
    && addgroup -S -g 1000 app \
    && adduser -S -D -H -u 1000 -G app app

WORKDIR /app

COPY --from=builder /out/db_backup ./db_backup

# Dumps are written to /app/dump (relative "dump" dir). A bind-mounted host
# directory must be writable by UID 1000.
RUN mkdir dump && chown app:app dump

USER app

CMD ["./db_backup"]
