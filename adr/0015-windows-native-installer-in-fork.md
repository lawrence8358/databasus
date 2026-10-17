# ADR-0015: Ship a native Windows installer from the fork alongside the Docker image

- **Status:** Accepted
- **Date:** 2026-10-18
- **Tags:** deployment, packaging, windows, fork

## Context

[ADR-0001](./0001-use-docker-as-base.md) makes the Docker image the only deploy unit and explicitly
refuses a second distribution channel. That works for the upstream audience, which already runs
Docker.

The `lawrence8358/databasus` fork serves government agencies that cannot follow it. Their machines
are Windows, installing Docker or WSL2 is forbidden, and software must arrive as a single installer
that works without network access. Today those agencies get no scheduled backups or monitoring for
their PostgreSQL databases at all.

The fork keeps merging upstream releases, so whatever we add has to survive those merges with as
few conflicts as possible.

## Decision

The fork ships a native Windows x64 installer in addition to the Docker image. It runs offline on
Windows Server 2016 or newer and Windows 10/11, registers Databasus as an auto-start Windows
Service, and serves the same Web UI as the Docker build. Upstream ADR-0001 still holds for the
Docker image; this ADR adds a second channel only in the fork.

The Windows build is deliberately narrow:

- **PostgreSQL logical backup and restore only.** MySQL, MariaDB, MongoDB, physical backups, PITR,
  the verification agent and the embedded physical restore scripts are unavailable. Engines and
  modes without client tools are reported as unavailable and hidden in the UI rather than deleted
  from the code, so upstream changes to them still merge.
- **One PostgreSQL client version.** Only the newest PostgreSQL client tools are bundled, unlike the
  per-major set of [ADR-0005](./0005-why-assets-binaries-instead-of-dbs-installation.md). A newer
  `pg_dump` reads every supported server; restoring into an older server may fail, and the restore
  screen warns about it.
- **A bundled metadata PostgreSQL pinned to one major version.** The installer carries its own
  PostgreSQL server for Databasus's internal state, listening on loopback only with a random
  password, started and stopped by the service. Upgrades of Databasus apply minor releases of it
  only. Moving to a new major is a separate, planned change that handles `pg_upgrade`. An advanced
  option lets an administrator point at an external metadata database instead.
- **Windows code kept apart.** Windows-only behaviour lives in files only Windows compiles and in a
  separate packaging directory. Upstream files change only where a seam is needed: configurable
  paths, per-engine availability and client tool path resolution. Linux and Docker behaviour stays
  as it is.

## Alternatives considered

- **Tell these users to run Docker Desktop or WSL2.** Not allowed on their machines, so it serves
  nobody.
- **Bundle every PostgreSQL client major, as the Docker image does.** Restores into old servers
  would always work, but the installer grows by several client sets for a case we can warn about.
- **Track the newest PostgreSQL major for the metadata database.** Every major bump needs a data
  conversion on the administrator's machine, unattended. Pinning the major makes upgrades a binary
  swap.
- **Support every engine and backup type on Windows.** Physical backups and the other engines each
  need their own Windows tooling and testing. The agencies need PostgreSQL logical backups, so the
  rest waits until someone needs it.
- **Contribute the Windows build upstream.** Upstream has chosen Docker only, so this stays in the
  fork.

## Consequences

### Positive

- Agencies that ban Docker can install Databasus from one offline `.exe` and run it like any other
  Windows service.
- Merging upstream releases stays cheap because Windows code sits in its own files.

### Negative

- The fork carries a second channel that upstream does not test. CI has to build the installer and
  run an install-to-uninstall smoke test on Windows, or an upstream merge can break it unnoticed.
- Windows users get a subset of Databasus. The README has to say so plainly.
- A major upgrade of the metadata PostgreSQL is manual work we still owe.
- The installer is not code-signed, so Windows shows an unknown-publisher warning.

## References

- [ADR-0001](./0001-use-docker-as-base.md) — the Docker-only decision this fork departs from.
- [ADR-0005](./0005-why-assets-binaries-instead-of-dbs-installation.md) — vendored client binaries,
  which the Windows build reuses with a single PostgreSQL version.
