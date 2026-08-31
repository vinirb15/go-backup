# Database Backup Scheduler

This Go project automates daily database backups for MySQL and PostgreSQL using cron jobs. The backups are stored in the `./dump` directory, and logs are recorded in `./dump/backup.log`.

## Features
- Supports MySQL and PostgreSQL
- Runs daily backups at midnight
- Stores backups in the `./dump` directory
- Logs errors and execution details in `./dump/backup.log`

## Requirements
- Go 1.22+
- MySQL or PostgreSQL client tools (`mysqldump`, `pg_dump`)
- Docker + Docker Compose (optional)

## Installation
### Clone the repository
```sh
git clone https://github.com/yourusername/backup-go.git
cd backup-go
```

### Initialize Go modules
```sh
go mod init backup-go
go mod tidy
```

## Configuration
Create a `.env` file in the project root with the following variables:
```ini
DB_TYPE=mysql  # or postgres
DB_HOST=127.0.0.1
DB_PORT=3306  # or 5432 for PostgreSQL
DB_NAME=your_database
DB_USER=username
DB_PASS=password
CRON_TIME=0 0 * * *  # optional, cron schedule (5 fields, robfig/cron/v3); defaults to midnight
TZ=America/Sao_Paulo  # optional, IANA timezone for the schedule; defaults to the system timezone
```

## Running the Backup Script
### Locally
```sh
go run main.go
```

### Using Docker Compose
Build and start the scheduler in the background:
```sh
docker compose up -d --build
```
- Config is read from `.env` (`env_file`); dumps are written to `./dump` on the host.
- View logs: `docker compose logs -f`
- Stop: `docker compose down`

If the database runs on the Docker host, set `DB_HOST=host.docker.internal` in
`.env` (not `localhost`, which would point at the container itself).

## Restoring a Backup
Use the `restore` subcommand to load a dump into the database described by the
current `DB_*` variables:
```sh
go run . restore dump/backup_<DB_NAME>_<TIMESTAMP>.sql
```
- MySQL dumps are plain SQL and are piped into the `mysql` client.
- PostgreSQL dumps use the custom format (`pg_dump -F c`) and are restored with
  `pg_restore --clean --if-exists --no-owner`.
- The target database must already exist (the dumps do not include `CREATE DATABASE`).

To restore into **another server**, point the `DB_*` variables at the target before
running — either edit `.env`, or override them in the environment (values already
exported take precedence over `.env`):
```sh
DB_HOST=new-host DB_NAME=new-db go run . restore dump/backup_app_20260101_000000.sql
```

With Docker Compose:
```sh
docker compose run --rm backup ./db_backup restore dump/<file>
```

## Logs & Backup Files
- Backups are saved in `./dump/backup_<DB_NAME>_<TIMESTAMP>.sql`
- Logs are recorded in `./dump/backup.log`

## License
MIT License

