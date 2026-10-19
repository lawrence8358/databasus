# Windows installer

`databasus.iss` is the Inno Setup script for the offline Windows x64 installer, and `build.ps1`
builds it. The installer needs no network access: it carries the backend, the launcher, the
migrations, the Web UI, the bundled PostgreSQL 18 (server and client tools) and the VC++ runtime
DLLs those binaries need.

## Building

Needs Go 1.26, swag, Node.js with pnpm, and Inno Setup 7.1.0 (`ISCC.exe` on `PATH`, in its default
install folder, or passed with `-IsccPath`).

```powershell
$env:VITE_APP_VERSION = '1.2.3'   # the version the Web UI shows; without it the UI says "vdev"
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
disk, that the backups folder follows the rules in [Changing the backups folder](#changing-the-backups-folder),
and that the port is free. It then registers and starts the `Databasus` service, waits until
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
| `/BACKUPSDIR=` | `<data dir>\backups` (local storage; see [Changing the backups folder](#changing-the-backups-folder) for the rules) |
| `/PORT=` | `4005` |
| `/EXTERNALDBFILE=` | none (bundled PostgreSQL); a file holding an external metadata database connection string |
| `/BUNDLEDDB` | switches an install that used an external database back to the bundled one |
| `/LANACCESS` | off (127.0.0.1 only); opens the Web UI to the local network over HTTPS only |
| `/PFX=` | with `/LANACCESS`, a PFX certificate file to import |
| `/PFXPASSWORDFILE=` | a file whose first line is the PFX password |
| `/SELFSIGNED` | with `/LANACCESS`, generates a new self-signed certificate |
| `/LOCALONLY` | switches an install with LAN access back to 127.0.0.1 |

### Changing the backups folder

The backups folder can change after installation, for example when drive C: runs out of space: in
the Web UI under **Databasus settings > Backups folder** (instance administrators only), or by
running Setup again with `/BACKUPSDIR=` or a new value on its folders page (see
[Upgrading](#upgrading)). New backups go to the new folder at once; the service does not restart.

Nothing is moved. Backups made earlier stay in the folder they were written to and remain
restorable, and retention still deletes them there, so the old disk frees up as they expire. The
settings page lists those earlier folders. To free the space at once, stop the service, move the
files from an earlier folder into the current one, and start it again: Databasus looks for a backup
in the current folder first, then in the earlier ones.

The folder must be an absolute path on a local fixed disk of this computer: no network path, mapped
network drive or removable drive. It must not be a drive root, the Windows folder, a Program Files
folder, `ProgramData`, `C:\Users` or `C:\Users\Public` themselves, the install folder or anything in it,
or anything in the program data folder other than its `backups` folder. A junction or symbolic
link that leads to such a place is refused too. Databasus creates a missing folder and checks that
the service account (LocalSystem) can write and delete a file in it. Each change is recorded in the
audit log.

The choice is stored in `<data dir>\backups-folder.txt` (readable and writable by SYSTEM and
Administrators only), together with the folders used before. It is the one place that decides where
backups go: the Web UI and Setup both write it, and the service command line does not carry the
folder, so a restart or an upgrade cannot undo a change. Docker and Linux installations do not show
this setting; their backups folder comes from the volume and `DATABASUS_BACKUPS_DIR` as before.
From a command line, with the service stopped:

```powershell
& "C:\Program Files\Databasus\databasus-launcher.exe" backups-folder -data-dir C:\ProgramData\Databasus             # shows it
& "C:\Program Files\Databasus\databasus-launcher.exe" backups-folder -data-dir C:\ProgramData\Databasus -set E:\Backups
```

### External metadata database

The wizard's "Metadata database (advanced)" page, or `/EXTERNALDBFILE=<file>` in silent mode, makes
Databasus keep its metadata on an existing PostgreSQL server. Setup tests the connection first and
installs nothing if it fails (exit code `7` when silent). The bundled PostgreSQL is then neither
initialized nor started. The connection string (`postgres://user:password@host:5432/db` or
`host=... user=... password=...`) is never accepted on the command line, because Setup logs its
command line. A typed string reaches the launcher through an environment variable of the Setup
process; either way the launcher stores it in `<data dir>\external-metadata-db.dsn` with an ACL
limited to SYSTEM and Administrators, and the service command line only names
that file. An upgrade keeps the choice; leave the field empty to keep the stored string.

### LAN access over HTTPS

By default the Web UI listens on `127.0.0.1` only. The wizard's "Network access" page, or
`/LANACCESS` in silent mode, lets other computers open it: the service then listens on every
interface (IPv4 and IPv6) and speaks HTTPS only on the Web UI port. There is no plain HTTP listener;
an `http://` request to that port gets `400 Client sent an HTTP request to an HTTPS server`. The
launcher refuses to start on a non-loopback address without HTTPS.

The certificate is one of:

- a PFX file, for example one issued by your organization (`/PFX=<file>`). Setup checks it before
  installing anything (wrong password, no private key, expired, or not for server authentication:
  exit code `7` when silent). The password is never accepted on the command line, because Setup logs
  its command line: type it on the page, or give `/PFXPASSWORDFILE=<file>`.
- a self-signed certificate Setup generates (`/SELFSIGNED`), valid for 5 years, for `localhost`, the
  computer name, its DNS name and its current IP addresses. Browsers warn until it is trusted; the
  Setup log shows its SHA-256 fingerprint so users can compare it.
- the certificate already stored, which is the default when it is still valid. Without one, silent
  mode generates a self-signed certificate.

Setup stores the certificate chain and key as PEM in `<data dir>\tls\databasus.crt` and
`databasus.key`, with an ACL limited to SYSTEM and Administrators, and adds
the inbound Windows Firewall rule `Databasus Web UI` for the port and `databasus.exe`, on the domain
and private profiles (add the public profile yourself if the server's network is classified public).
`/LOCALONLY`, or unchecking the option on a rerun, removes the rule; the uninstaller removes it too.
To replace an expiring certificate, run Setup again with `/LANACCESS /PFX=...` or `/SELFSIGNED`.

Exit codes: `0` installed and the Web UI answers; `1` unsupported system or invalid folder;
`7` a check failed before anything was installed; `10` the files are installed but the service did
not register or start. The reason is in the `/LOG` file, and for `10` also in
`<data dir>\logs\launcher.log` and the Application event log (source `Databasus`).

## Upgrading

Run the new version's `.exe` over the installed one, interactively or with the same silent command
line minus the folder options. Setup reads the earlier choices from `HKLM\SOFTWARE\Databasus` and
keeps them without asking again: the program data folder and the Web UI port (`/DATADIR` and `/PORT`
are ignored on an upgrade, because moving the data folder would start the service on an empty
metadata database), the metadata database choice (only `/EXTERNALDBFILE` or `/BUNDLEDDB` change
it), and LAN access and its certificate (only `/LANACCESS`, `/LOCALONLY`, `/PFX` or `/SELFSIGNED`
change them). The Ready page lists what is kept.

The backups folder is the one the service uses now, read from `<data dir>\backups-folder.txt`, so a
change made in the Web UI is kept. An upgrade can change it: `/BACKUPSDIR=` in silent mode, or the
folders page, which an interactive upgrade shows with the program data folder locked. A new folder
is checked by the rules in [Changing the backups folder](#changing-the-backups-folder) before the
service is stopped (exit code `7` when silent); the folder kept as it is is not checked again.
Backups in the earlier folder stay there and remain restorable.

Before it touches anything, Setup checks the existing metadata database: if it was created by
another PostgreSQL major version than the one the new version bundles, Setup stops with the reason
(exit code `7` when silent) and leaves the running service and its data as they are. Moving to a
new major needs `pg_upgrade`, which Setup does not do. Within one major (for example 18.4 to 18.6)
the new binaries simply run the existing data.

Setup also refuses to install an older version over a newer one (exit code `7` when silent), because
the newer version may already have migrated the metadata database. Versions compare by
major.minor.patch. Development and CI builds (`0.0.0-<suffix>`) skip this check.

Otherwise Setup stops the service, replaces the program files (the `migrations`, `ui\build` and
bundled PostgreSQL folders as a whole, so nothing of the old version lingers), updates the service
registration and starts it again. The backend applies the new metadata migrations at that start,
and Setup waits until the Web UI answers. Users, databases, schedules, settings, the secret key and
all backups stay.

## Uninstalling

From Apps & features, or `"<install dir>\unins000.exe"`. The program data folder and the backups
folders stay on disk unless you choose to delete all data, which asks a second time because without
the secret key encrypted settings and backups cannot be recovered. The program data folder is
deleted as a whole. The backups folders (the current one and every earlier one listed in
`backups-folder.txt`) may be shared with other programs, so only the files Databasus named there
(backups, their sidecars and WAL segments, plus its own `temp` and `wal-queue` entries) are deleted,
and each folder itself only when nothing else is left in it. Silently:

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

`e2e/windows-installer/upgrade.ps1` covers an upgrade: it installs the previous release, creates a
user, a workspace, a local storage, a database, a backup and a weekly schedule, checks that the
current installer refuses a metadata database of another PostgreSQL major without stopping the
service, installs the current version with no folder or port options, and then checks the version
changed, the service kept its folders and port, the metadata database is at the newest shipped
migration, everything created before is still there, the old backup restores with the same rows,
and a new backup works. CI runs it in the `test-windows-installer-upgrade` job, which builds the
previous installer from the newest `v*` tag that contains `packaging/windows/databasus.iss` (and
only warns while there is none). By hand, with the same machine requirements:

```powershell
pwsh -File e2e/windows-installer/upgrade.ps1 `
    -PreviousInstallerPath databasus-setup-1.2.2-x64.exe -InstallerPath databasus-setup-1.2.3-x64.exe
```

It uses port 4124 for the Web UI and 5498 for the throwaway source server, keeps backups in
`<data dir>\backups`, and leaves the program data folder behind like `run.ps1`.

`e2e/windows-installer/lan-access.ps1` checks the LAN access option: an install with a self-signed
certificate listens on every interface, answers HTTPS on the computer's network address and not
plain HTTP, has the firewall rule and a restricted private key; a rerun with `/PFX` serves the
imported certificate, a rerun without options keeps it, `/LOCALONLY` goes back to `127.0.0.1` and
removes the rule, and the uninstaller removes the rule. It needs an elevated PowerShell 7, uses port
4125 and `C:\ProgramData\Databasus-lan-test`, and deletes that folder at the end.

```powershell
pwsh -File e2e/windows-installer/lan-access.ps1 -InstallerPath databasus-setup-1.2.3-x64.exe
```
