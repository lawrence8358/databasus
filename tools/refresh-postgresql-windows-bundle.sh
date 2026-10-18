#!/usr/bin/env bash
# Rebuild the Windows x64 PostgreSQL client bundle under
# assets/tools/win-x64/postgresql/ from EDB's official binaries archive.
#
#   tools/refresh-postgresql-windows-bundle.sh
#
# The Windows build ships one client, from the newest supported major, and
# serves every supported server with it. The release and the archive checksum
# are pinned below rather than resolved as "the newest patch", so a refresh
# gives the same bytes whatever day it runs. To move to a new patch, change
# PG_RELEASE and ARCHIVE_SHA256 together.
#
# Only the logical backup clients are copied, with the DLLs they load. The
# list was taken from the import tables of the three programs; the Microsoft
# Visual C++ runtime (VCRUNTIME140.dll) is the one dependency left to the host,
# because EDB does not ship it in this archive.
set -euo pipefail

PG_MAJOR="18"
PG_RELEASE="18.6-1"
ARCHIVE_SHA256="fbe23da234ee31547bf8a36d29dfd81e82b849df2d2b78d2eecb43d360252f8c"
ARCHIVE_URL="https://get.enterprisedb.com/postgresql/postgresql-${PG_RELEASE}-windows-x64-binaries.zip"

BUNDLE_FILES=(
  pg_dump.exe
  pg_restore.exe
  psql.exe
  libpq.dll
  libssl-3-x64.dll
  libcrypto-3-x64.dll
  libintl-9.dll
  libiconv-2.dll
  libwinpthread-1.dll
  liblz4.dll
  libzstd.dll
)

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$REPO_ROOT/assets/tools/win-x64/postgresql/postgresql-$PG_MAJOR/bin"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "==> $ARCHIVE_URL"
curl -fsSL -o "$WORK/binaries.zip" "$ARCHIVE_URL"
echo "$ARCHIVE_SHA256  $WORK/binaries.zip" | sha256sum -c -

archive_members=()
for file in "${BUNDLE_FILES[@]}"; do
  archive_members+=("pgsql/bin/$file")
done

unzip -q "$WORK/binaries.zip" "${archive_members[@]}" -d "$WORK/unpacked"

rm -rf "$DEST"
mkdir -p "$DEST"
for file in "${BUNDLE_FILES[@]}"; do
  cp "$WORK/unpacked/pgsql/bin/$file" "$DEST/$file"
done

echo "==> refreshed $DEST"
ls -l "$DEST"
