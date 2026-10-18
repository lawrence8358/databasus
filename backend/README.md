# Before run

> Copy .env.example to .env in the repo root (single source for backend, frontend, and docker compose)

# Run

Commands run from the repository root, with the tools from `mise.toml` installed.

To run the backend together with the frontend, the dev database, the test fixtures and the WAL backup source:

> just run

To run tests:

> just test

Before commit:

> just lint

# Migrations

The backend applies pending migrations on startup. To create a migration:

> goose -dir backend/migrations create MIGRATION_NAME sql

To apply or roll back migrations on the dev database, with the variables from `.env` exported (goose reads `GOOSE_DRIVER` and `GOOSE_DBSTRING` from the environment):

> goose -dir backend/migrations up

> goose -dir backend/migrations down

# Swagger

`just run` generates the swagger docs before the backend starts. To regenerate them by hand, from `backend/`:

> swag init -d . -g cmd/main.go -o swagger

Swagger URL is:

> http://localhost:4005/api/v1/docs/swagger/index.html#/

# Run natively on Windows

The backend also builds as a Windows x64 executable, for the fork's Windows installer. It runs
PostgreSQL logical backups only; the other engines and the physical backup features report as
unavailable. From `backend/`, after generating the swagger docs:

> GOOS=windows GOARCH=amd64 go build -o databasus.exe ./cmd

Put `migrations/`, `assets/tools/win-x64/` (as `assets/tools/win-x64/`) and, for the Web UI, the
built frontend (as `ui/build/`) next to `databasus.exe`; it finds them relative to its own location,
not the working directory. Then point it at an existing PostgreSQL for metadata and start it:

> set ENV_MODE=production
>
> set DATABASE_DSN=host=127.0.0.1 port=5432 user=postgres password=... dbname=databasus sslmode=disable
>
> set DATABASUS_DATA_DIR=C:\ProgramData\Databasus
>
> set DATABASUS_BACKUPS_DIR=D:\DatabasusBackups
>
> databasus.exe

The client tools need the Microsoft Visual C++ runtime (`VCRUNTIME140.dll`), which most Windows
machines already have.
