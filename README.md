# Database Backup Scheduler

This Go project automates daily database backups for MySQL and PostgreSQL using cron jobs. The backups are stored in the `./dump` directory, and logs are recorded in `./dump/backup.log`.

## Features
- Supports MySQL and PostgreSQL
- Runs daily backups at midnight
- Stores backups in the `./dump` directory
- Logs errors and execution details in `./dump/backup.log`

## Requirements
- Go 1.20+
- MySQL or PostgreSQL client tools (`mysqldump`, `pg_dump`)
- Docker (optional)

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
```

## Running the Backup Script
### Locally
```sh
go run main.go
```

### Using Docker
#### Build the image
```sh
docker build -t db-backup .
```

#### Run the container
```sh
docker run --env-file .env -v $(pwd)/dump:/root/dump db-backup
```

## Logs & Backup Files
- Backups are saved in `./dump/backup_<DB_NAME>_<TIMESTAMP>.sql`
- Logs are recorded in `./dump/backup.log`

## License
MIT License

