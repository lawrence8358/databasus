---
name: windows-installer-release
description: Build, verify and publish the Databasus Windows x64 installer. Use when asked to build the Windows installer or Windows version, run its smoke tests on a real machine, or publish a Windows GitHub release.
---

# Windows installer: build, verify, release

`packaging/windows/README.md` is the source of truth for the build commands, installer options and smoke tests. This skill adds what an agent needs on top: toolchain setup, version rules, how to run the elevated tests, and the release steps.

## 1. Toolchain

The build needs Go (version in `mise.toml`), `swag` (same), Node.js with pnpm, and Inno Setup 7.1.0. On a Windows host without them, use portable copies outside the repo instead of system installs:

- Go: unzip `https://go.dev/dl/go<version>.windows-amd64.zip` and prepend its `bin` to `PATH`.
- swag: `go install github.com/swaggo/swag/cmd/swag@v<version>` with `GOBIN` pointing at a folder on `PATH`.
- Inno Setup: pass `-IsccPath <folder>\ISCC.exe` to `build.ps1` when it is not installed.

Done when `go version`, `swag --version`, `pnpm --version` and `ISCC.exe /?` all answer.

## 2. Version

Pick the version before building; it goes into `VITE_APP_VERSION` and `build.ps1 -Version`, and both must match.

- A release is `<upstream version>-windows.<n>`, for example `3.61.0-windows.1`: the upstream tag the branch is based on (`git describe --tags --abbrev=0 upstream/main`, where `upstream` is `databasus/databasus`), plus a counter for fork-only rebuilds of the same upstream version.
- A test build is `0.0.0-<suffix>`. Setup skips the downgrade check for it, so it installs over a release.
- Setup compares only `major.minor.patch` and refuses a lower one over a higher one, so a release version never goes below an earlier release.

## 3. Build

Run the `## Building` commands from `packaging/windows/README.md` with that version. Done when `packaging/windows/out/databasus-setup-<version>-x64.exe` exists (about 76 MB) and its file version matches.

## 4. Verify on a real machine

Installing needs admin rights. On a workstation where the user is in Administrators behind UAC, elevate with `Start-Process pwsh -Verb RunAs -Wait`, and bundle every elevated step into one script per round that writes its own transcript and results file, because each round costs the user one UAC click. A UAC prompt only gets answered while the user is at an unlocked screen: tell the user before launching a round, and check that the round actually started (its log exists) before waiting on it. A non-elevated process cannot send input to elevated windows, so UI automation of the installer or MMC runs inside the elevated script.

Run, in this order, and record pass or fail with the log path for each:

1. `e2e/windows-installer/run.ps1`: silent install, backup and restore on a second disk, restart, uninstall.
2. `e2e/windows-service/run.ps1`: service install, restart, recovery after a forced kill, event log, uninstall.
3. `e2e/windows-installer/lan-access.ps1`: HTTPS, firewall rule.
4. `e2e/windows-installer/upgrade.ps1` against the previous release installer, from the newest `v*-windows.*` release.

Done when all four pass on the exact installer you will publish, and the machine is clean afterwards: Databasus uninstalled, test containers, temporary users, VHDs and firewall rules removed.

## 5. Release

Publish only an installer that passed step 4. The fork's releases live on `lawrence8358/databasus`; release notes and the install guide are written in Traditional Chinese for the agencies that use this build.

1. Tag the commit the installer was built from: `git tag v<version> <commit>` and push the tag.
2. Write the notes to a file: what changed since the previous Windows release, system requirements, install, upgrade and uninstall steps, and known limitations. Base the steps on `packaging/windows/README.md`.
3. `gh release create v<version> packaging/windows/out/databasus-setup-<version>-x64.exe --repo lawrence8358/databasus --title "Databasus <version>（Windows）" --notes-file <notes>`, plus a `SHA256SUMS.txt` asset made with `Get-FileHash -Algorithm SHA256`.

Done when `gh release view v<version> --repo lawrence8358/databasus` lists both assets and the downloaded installer's hash matches `SHA256SUMS.txt`.
