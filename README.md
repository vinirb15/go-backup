# Database Backup Scheduler

This Go project automates daily database backups for MySQL and PostgreSQL using cron jobs. The backups are stored in the `./dump` directory, and logs are recorded in `./dump/backup.log`.

## Features
- Supports MySQL and PostgreSQL
- Backs up multiple independent databases (each with its own type, host and credentials)
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
Create a `.env` file in the project root (see `env-example`).

### Multiple databases
List the database aliases in `DATABASES`; each alias reads its own
`<ALIAS>_DB_*` variables (alias in uppercase). All databases are backed up, one
after another, on every scheduled run — a failure in one does not stop the others.
```ini
CRON_TIME=0 0 * * *  # optional, cron schedule (5 fields, robfig/cron/v3); defaults to midnight
TZ=America/Sao_Paulo  # optional, IANA timezone for the schedule; defaults to the system timezone

DATABASES=app,crm

APP_DB_TYPE=postgres
APP_DB_HOST=127.0.0.1
APP_DB_PORT=5432
APP_DB_NAME=app
APP_DB_USER=username
APP_DB_PASS=password

CRM_DB_TYPE=mysql
CRM_DB_HOST=10.0.0.5
CRM_DB_PORT=3306
CRM_DB_NAME=crm
CRM_DB_USER=username
CRM_DB_PASS=password
```
Aliases may contain only letters, numbers and `_`. On startup every missing
variable is reported by name (e.g. `CRM_DB_PASS`).

### Single database
Without `DATABASES`, a single database is read from the unprefixed variables:
```ini
DB_TYPE=mysql  # or postgres
DB_HOST=127.0.0.1
DB_PORT=3306  # or 5432 for PostgreSQL
DB_NAME=your_database
DB_USER=username
DB_PASS=password
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
configuration of the given alias:
```sh
go run . restore <alias> dump/backup_<alias>_<TIMESTAMP>.sql
```
With a single database configured (including the unprefixed `DB_*` mode) the
alias can be omitted:
```sh
go run . restore dump/backup_<DB_NAME>_<TIMESTAMP>.sql
```
- MySQL dumps are plain SQL and are piped into the `mysql` client.
- PostgreSQL dumps use the custom format (`pg_dump -F c`) and are restored with
  `pg_restore --clean --if-exists --no-owner`.
- The target database must already exist (the dumps do not include `CREATE DATABASE`).

To restore into **another server**, point the alias' `DB_*` variables at the target before
running — either edit `.env`, or override them in the environment (values already
exported take precedence over `.env`):
```sh
APP_DB_HOST=new-host APP_DB_NAME=new-db go run . restore app dump/backup_app_20260101_000000.sql
```

With Docker Compose:
```sh
docker compose run --rm backup ./db_backup restore <alias> dump/<file>
```

## Logs & Backup Files
- Backups are saved in `./dump/backup_<alias>_<TIMESTAMP>.sql` (`<DB_NAME>` in
  single-database mode); failed dumps are removed
- Logs are recorded in `./dump/backup.log`

## License
MIT License

