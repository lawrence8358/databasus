# Windows installer

`databasus.iss` is the Inno Setup script for the offline Windows x64 installer, and `build.ps1`
builds it. The installer needs no network access: it carries the backend, the launcher, the
migrations, the Web UI, the bundled PostgreSQL 18 (server and client tools) and the VC++ runtime
DLLs those binaries need.

## Building

Needs Go 1.26, swag, Node.js with pnpm, and Inno Setup 7.1.0 (`ISCC.exe` on `PATH`, in its default
install folder, or passed with `-IsccPath`).

```powershell
cd frontend; pnpm install --frozen-lockfile; pnpm build; cd ..
cd backend; swag init -d . -g cmd/main.go -o swagger; cd ..
pwsh packaging/windows/build.ps1 -Version 1.2.3
```

The installer lands in `packaging/windows/out/databasus-setup-<version>-x64.exe`.

The VC++ runtime (`vcruntime140.dll`, `vcruntime140_1.dll`, `msvcp140.dll`) is copied next to the
PostgreSQL binaries from the Visual Studio redistributable folder, or from `System32` when Visual
Studio is not installed. Windows 10 and Server 2016 already include the Universal CRT.

## Installing

Run the `.exe` as an administrator. Setup stops with the reason on anything older than Windows
Server 2016 / Windows 10 version 1607, and on anything but 64-bit Windows for x64 processors. It
asks for the install folder, the program data folder, the backups folder and the Web UI port, and
checks that the LocalSystem service account can write to both folders, that they are on a local
disk, and that the port is free. It then registers and starts the `Databasus` service, waits until
the Web UI answers on `http://127.0.0.1:<port>/`, and offers to open it.

Silent install, where every option has a parameter:

```powershell
databasus-setup-1.2.3-x64.exe /VERYSILENT /SUPPRESSMSGBOXES /NORESTART /LOG=C:\setup.log `
    /DIR="C:\Program Files\Databasus" /DATADIR=C:\ProgramData\Databasus /BACKUPSDIR=D:\DatabasusBackups /PORT=4005
```

| Parameter | Default |
|---|---|
| `/DIR=` | `C:\Program Files\Databasus` |
| `/DATADIR=` | `C:\ProgramData\Databasus` (metadata database, secret key, logs) |
| `/BACKUPSDIR=` | `<data dir>\backups` (local storage; a local disk only) |
| `/PORT=` | `4005` |

Exit codes: `0` installed and the Web UI answers; `1` unsupported system or invalid folder;
`7` a check failed before anything was installed; `10` the files are installed but the service did
not register or start. The reason is in the `/LOG` file, and for `10` also in
`<data dir>\logs\launcher.log` and the Application event log (source `Databasus`).

## Uninstalling

From Apps & features, or `"<install dir>\unins000.exe"`. The program data folder and the backups
folder stay on disk unless you choose to delete all data, which asks a second time because without
the secret key encrypted settings and backups cannot be recovered. Silently:

```powershell
& "C:\Program Files\Databasus\unins000.exe" /VERYSILENT /SUPPRESSMSGBOXES             # keeps the data
& "C:\Program Files\Databasus\unins000.exe" /VERYSILENT /SUPPRESSMSGBOXES /DELETEDATA # deletes it
```

## Smoke test

`e2e/windows-installer/run.ps1` installs silently, checks the service and that the Web UI listens
on `127.0.0.1` only, creates a user, a workspace, a local storage and a PostgreSQL database through
the API, backs it up into the backups folder on a second disk, restores it into another database
and compares the rows, restarts the service and checks the state is kept, then uninstalls silently
and checks the data and backups are still there. CI runs it in the `test-windows-installer` job.

To run it by hand, use an elevated PowerShell 7 on a machine (or VM) without Databasus installed:

```powershell
pwsh -File e2e/windows-installer/run.ps1 -InstallerPath packaging\windows\out\databasus-setup-1.2.3-x64.exe
```

It creates a 2 GB VHD as the second disk (pass `-BackupsDir D:\some\folder` to use an existing
disk instead), runs a throwaway PostgreSQL server from the installed bundle on port 5499
(`-SourcePostgresPort`), and uses port 4123 for the Web UI (`-Port`). Afterwards the program data
folder (`C:\ProgramData\Databasus` unless `-DataDir` is given) is left behind on purpose; delete it
before the next run.
