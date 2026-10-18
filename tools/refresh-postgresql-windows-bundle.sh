#!/usr/bin/env bash
# Rebuild the Windows x64 PostgreSQL bundle under
# assets/tools/win-x64/postgresql/ from EDB's official binaries archive.
#
#   tools/refresh-postgresql-windows-bundle.sh
#
# The Windows build ships one PostgreSQL release, from the newest supported
# major. Its client tools serve every supported server, and its server runs
# the bundled metadata database. The release and the archive checksum are
# pinned below rather than resolved as "the newest patch", so a refresh gives
# the same bytes whatever day it runs. To move to a new patch, change
# PG_RELEASE and ARCHIVE_SHA256 together. Moving to a new major is not a
# refresh: the metadata database's data directory needs pg_upgrade then.
#
# Only what the backend and the launcher run is copied, with the DLLs those
# programs load. The DLL lists were taken from the import tables of the
# programs and of the server's loadable modules; the Microsoft Visual C++
# runtime (VCRUNTIME140.dll, VCRUNTIME140_1.dll, MSVCP140.dll) is the one
# dependency left to the host, because EDB does not ship it in this archive.
set -euo pipefail

PG_MAJOR="18"
PG_RELEASE="18.6-1"
ARCHIVE_SHA256="fbe23da234ee31547bf8a36d29dfd81e82b849df2d2b78d2eecb43d360252f8c"
ARCHIVE_URL="https://get.enterprisedb.com/postgresql/postgresql-${PG_RELEASE}-windows-x64-binaries.zip"

CLIENT_BIN_FILES=(
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

# The metadata database server. pg_ctl stops it and pg_resetwal is the
# launcher's last resort when a crashed cluster will not start.
SERVER_BIN_FILES=(
  postgres.exe
  initdb.exe
  pg_ctl.exe
  pg_resetwal.exe
  icudt77.dll
  icuin77.dll
  icuuc77.dll
  libxml2.dll
  zlib1.dll
)

# initdb loads PL/pgSQL and the snowball dictionaries; the encoding
# conversions are loaded on demand by any client using a non-UTF8 encoding.
# Translations (share/locale) and the optional extensions are left out.
# Extended regular expressions over archive member names, because unzip's
# wildcards do not cross directory levels in every build.
SERVER_MEMBER_PATTERNS=(
  '^pgsql/lib/(plpgsql|dict_snowball|[a-z0-9_]+_and_[a-z0-9_]+)\.dll$'
  '^pgsql/share/[^/]+\.(bki|sql|txt|sample)$'
  '^pgsql/share/extension/plpgsql[^/]*$'
  '^pgsql/share/(timezone|timezonesets|tsearch_data)/.*[^/]$'
)

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
DEST="$REPO_ROOT/assets/tools/win-x64/postgresql/postgresql-$PG_MAJOR"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

echo "==> $ARCHIVE_URL"
curl -fsSL -o "$WORK/binaries.zip" "$ARCHIVE_URL"
echo "$ARCHIVE_SHA256  $WORK/binaries.zip" | sha256sum -c -

archive_members=()
for file in "${CLIENT_BIN_FILES[@]}" "${SERVER_BIN_FILES[@]}"; do
  archive_members+=("pgsql/bin/$file")
done
while IFS= read -r member; do
  archive_members+=("$member")
done < <(zipinfo -1 "$WORK/binaries.zip" | grep -E "$(IFS='|'; echo "${SERVER_MEMBER_PATTERNS[*]}")")

unzip -q "$WORK/binaries.zip" "${archive_members[@]}" -d "$WORK/unpacked"

rm -rf "$DEST"
mkdir -p "$DEST/bin"
for file in "${CLIENT_BIN_FILES[@]}" "${SERVER_BIN_FILES[@]}"; do
  cp "$WORK/unpacked/pgsql/bin/$file" "$DEST/bin/$file"
done
cp -R "$WORK/unpacked/pgsql/lib" "$DEST/lib"
cp -R "$WORK/unpacked/pgsql/share" "$DEST/share"

echo "==> refreshed $DEST"
du -sh "$DEST/bin" "$DEST/lib" "$DEST/share"
