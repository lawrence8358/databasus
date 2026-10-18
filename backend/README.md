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
unavailable. A second program, the launcher, takes the place of the Docker image's entrypoint: it
runs the bundled PostgreSQL server for the metadata database and starts the backend against it.
From `backend/`, after generating the swagger docs:

> GOOS=windows GOARCH=amd64 go build -o databasus.exe ./cmd
>
> GOOS=windows GOARCH=amd64 go build -o databasus-launcher.exe ./cmd/launcher

Put `migrations/`, `assets/tools/win-x64/` (as `assets/tools/win-x64/`) and, for the Web UI, the
built frontend (as `ui/build/`) next to both executables; they find them relative to their own
location, not the working directory. Then start the launcher with a program data folder:

> databasus-launcher.exe -data-dir C:\ProgramData\Databasus -backups-dir D:\DatabasusBackups

On the first start it creates the metadata cluster in `<data-dir>\pgdata` with a random password,
kept in `<data-dir>\metadata-db.password`. The server listens on `127.0.0.1:5437` (change it with
`-metadata-db-port`) and accepts only password logins from that address; its log goes to
`<data-dir>\logs`. The launcher writes `ui\build\runtime-config.js`, passes the connection string
and the folders to `databasus.exe`, and on Ctrl+C stops the backend first and the database second.
If the database server was killed, the next start lets PostgreSQL recover it; if it cannot start at
all, the launcher resets its WAL once, as the Docker image does. `-backups-dir` defaults to
`<data-dir>\backups`; the other backend settings, such as `SMTP_HOST`, are read from the
environment as usual.

To run Databasus as a Windows service instead, register it from an administrator prompt with the
same flags; they are stored, as absolute paths, in the service's command line:

> databasus-launcher.exe install -data-dir C:\ProgramData\Databasus -backups-dir D:\DatabasusBackups
>
> sc.exe start Databasus

The `Databasus` service runs as LocalSystem, starts at boot, and is restarted by Windows 10, 30
and 60 seconds after it crashes or stops with an error. Running `install` again updates the
registration. Stopping the service stops the backend and then the database, as Ctrl+C does. The
launcher's own log and the console output of the backend and the database server go to
`<data-dir>\logs\launcher.log`; start and stop, and the reason the service stopped on its own,
also go to the Application event log under the `Databasus` source.
`databasus-launcher.exe uninstall` stops and removes the service and leaves the data folder alone.

To run the backend alone against a PostgreSQL you manage yourself, skip the launcher:

> set ENV_MODE=production
>
> set DATABASE_DSN=host=127.0.0.1 port=5432 user=postgres password=... dbname=databasus sslmode=disable
>
> set DATABASUS_DATA_DIR=C:\ProgramData\Databasus
>
> set DATABASUS_BACKUPS_DIR=D:\DatabasusBackups
>
> databasus.exe

The bundled PostgreSQL programs need the Microsoft Visual C++ runtime (`VCRUNTIME140.dll`,
`VCRUNTIME140_1.dll`, `MSVCP140.dll`), which most Windows machines already have.
