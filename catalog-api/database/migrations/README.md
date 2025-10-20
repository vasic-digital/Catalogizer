# Database Migrations

This directory contains database migrations for Catalogizer.

## Migration Files

Migrations are stored as SQL files with the following naming convention:
```
{version}_{name}.{direction}.sql
{version}_{name}.{database}.{direction}.sql
```

Examples:
- `000001_initial_schema.up.sql` - PostgreSQL migration (up)
- `000001_initial_schema.down.sql` - PostgreSQL rollback (down)
- `000001_initial_schema.sqlite.up.sql` - SQLite-specific migration (up)

## Running Migrations

### Using golang-migrate CLI

Install the CLI tool:
```bash
go install -tags 'postgres sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

Run migrations:
```bash
# PostgreSQL
migrate -path database/migrations -database "postgres://catalogizer:password@localhost:5432/catalogizer?sslmode=disable" up

# SQLite
migrate -path database/migrations -database "sqlite3://./catalogizer.db" up
```

Rollback:
```bash
migrate -path database/migrations -database "postgres://..." down 1
```

### Using the Application

Migrations run automatically when the application starts:
```bash
go run main.go
```

### Using Docker

Migrations run automatically when the Docker container starts.

## Creating New Migrations

1. Determine the next version number (increment the last migration)
2. Create both up and down migration files:
   ```bash
   # For PostgreSQL
   touch database/migrations/000002_add_users.up.sql
   touch database/migrations/000002_add_users.down.sql

   # For SQLite (if syntax differs)
   touch database/migrations/000002_add_users.sqlite.up.sql
   touch database/migrations/000002_add_users.sqlite.down.sql
   ```

3. Write the SQL for your changes in the `.up.sql` file
4. Write the rollback SQL in the `.down.sql` file

## Migration Guidelines

1. **Always create both up and down migrations**
2. **Test migrations on both PostgreSQL and SQLite**
3. **Use transactions where possible**
4. **Make migrations idempotent** (use IF NOT EXISTS, etc.)
5. **Never modify existing migrations** that have been deployed
6. **Document complex migrations** with comments
7. **Test rollbacks** before deploying

## Database-Specific Syntax

### PostgreSQL vs SQLite Differences

| Feature | PostgreSQL | SQLite |
|---------|------------|--------|
| Auto-increment | SERIAL | INTEGER PRIMARY KEY AUTOINCREMENT |
| Boolean | BOOLEAN | INTEGER (0/1) |
| Timestamp | TIMESTAMP | DATETIME |
| Current time | CURRENT_TIMESTAMP | CURRENT_TIMESTAMP |
| Big integers | BIGINT | INTEGER |

When creating database-specific migrations, use the `.postgres.up.sql` or `.sqlite.up.sql` suffix.

## Migration Status

Check which migrations have been applied:
```bash
migrate -path database/migrations -database "..." version
```

## Troubleshooting

### Dirty Database State

If a migration fails mid-way, the database may be in a "dirty" state:
```bash
migrate -path database/migrations -database "..." force VERSION
```

### Reset Database (Development Only)

To reset the database to a clean state:
```bash
migrate -path database/migrations -database "..." drop
migrate -path database/migrations -database "..." up
```

**WARNING**: This will delete all data!

## Migration History

| Version | Name | Description |
|---------|------|-------------|
| 000001 | initial_schema | Creates base tables for storage roots, files, metadata, etc. |
