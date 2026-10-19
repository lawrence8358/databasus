# End-to-end smoke test of an in-place upgrade with the offline Windows installer, driven only
# through the installer command lines, the Windows service and the HTTP API (plus a read of the
# metadata database's migration table):
#
#   silent install of the previous release -> create a user, a workspace, a local storage, a local
#   PostgreSQL database and a backup schedule -> back up -> Setup of the current version refuses a
#   metadata database of another PostgreSQL major and leaves the running service alone -> silent
#   install of the current version with no folder or port options -> same folders and port, newest
#   migration applied, everything still there -> restore the old backup and compare the data ->
#   back up again with the new version.
#
# Needs an elevated PowerShell 7 on a machine without Databasus installed. The source database runs
# from a copy of the previous release's PostgreSQL bundle, because the upgrade replaces the
# installed one. How to run it by hand is in packaging/windows/README.md.
param(
    [Parameter(Mandatory = $true)][string]$PreviousInstallerPath,
    [Parameter(Mandatory = $true)][string]$InstallerPath,
    [string]$InstallDir = (Join-Path $env:ProgramFiles 'Databasus'),
    [string]$DataDir = (Join-Path $env:ProgramData 'Databasus'),
    [string]$BackupsDir = '',
    [int]$Port = 4124,
    [int]$SourcePostgresPort = 5498,
    [int]$MetadataDbPort = 5437,
    [string]$WorkDir = (Join-Path ([System.IO.Path]::GetTempPath()) 'databasus-installer-upgrade')
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$serviceName = 'Databasus'
$baseUrl = "http://127.0.0.1:$Port/api/v1"
$previousInstallerPath = (Resolve-Path $PreviousInstallerPath).Path
$installerPath = (Resolve-Path $InstallerPath).Path
if (-not $BackupsDir) { $BackupsDir = Join-Path $DataDir 'backups' }
$sourcePassword = 'upgrade-source-password'
$userEmail = 'upgrade@example.com'
$userPassword = 'Upgrade-Passw0rd!'
$previousSetupLog = Join-Path $WorkDir 'setup-previous.log'
$refusedSetupLog = Join-Path $WorkDir 'setup-refused.log'
$upgradeSetupLog = Join-Path $WorkDir 'setup-upgrade.log'
$setupLog = $previousSetupLog
$uninstallLog = Join-Path $WorkDir 'uninstall.log'
$sourceDataDir = Join-Path $WorkDir 'source-pgdata'
$vhdPath = Join-Path $WorkDir 'unused.vhdx'

. (Join-Path $PSScriptRoot 'common.ps1')

function Invoke-Setup([string]$Path, [string]$LogPath, [string[]]$Options = @()) {
    $script:setupLog = $LogPath
    $arguments = @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', "/LOG=`"$LogPath`"") + $Options
    (Start-Process -FilePath $Path -Wait -PassThru -ArgumentList $arguments).ExitCode
}

function Get-AppVersion { (Invoke-Api GET '/system/version').version }

function Get-BackupById([string]$DatabaseId, [string]$BackupId, [string]$Token) {
    @((Invoke-Api GET "/backups?database_id=$DatabaseId" $null $Token).backups) | Where-Object id -eq $BackupId
}

function Wait-ForBackup([string]$DatabaseId, [string]$Token) {
    Invoke-Api POST '/backups' @{ database_id = $DatabaseId } $Token | Out-Null
    $status = Wait-ForStatus {
        $latest = @((Invoke-Api GET "/backups?database_id=$DatabaseId&limit=1" $null $Token).backups)
        if ($latest.Count -gt 0) { $latest[0].status } else { '' }
    } 'the backup to finish'
    $backup = @((Invoke-Api GET "/backups?database_id=$DatabaseId&limit=1" $null $Token).backups)[0]
    if ($status -ne 'COMPLETED') { throw "the backup ended $status`: $($backup.failMessage)" }
    return $backup
}

function Get-ScheduleSummary($BackupConfig) {
    $interval = $BackupConfig.backupInterval
    "enabled=$($BackupConfig.isBackupsEnabled) type=$($interval.type) time=$($interval.timeOfDay) " +
        "weekday=$($interval.weekday) storage=$($BackupConfig.storage.id)"
}

# Reads the migration table of the bundled metadata database with the password file the launcher
# keeps in the data folder (readable by administrators).
function Get-AppliedMigrationVersion {
    $env:PGPASSWORD = (Get-Content (Join-Path $DataDir 'metadata-db.password') -Raw).Trim()
    try {
        $output = & (Join-Path (Get-PostgresqlBin) 'psql.exe') -h 127.0.0.1 -p $MetadataDbPort -U postgres -d databasus `
            -v ON_ERROR_STOP=1 -X -A -t -c 'SELECT max(version_id) FROM goose_db_version WHERE is_applied'
        if ($LASTEXITCODE -ne 0) { throw 'reading goose_db_version from the metadata database failed' }
    } finally {
        $env:PGPASSWORD = $sourcePassword
    }
    [long](($output -join '').Trim())
}

function Get-NewestShippedMigrationVersion {
    $versions = Get-ChildItem (Join-Path $InstallDir 'migrations') -Filter '*.sql' |
        ForEach-Object { if ($_.Name -match '^(\d+)_') { [long]$Matches[1] } }
    ($versions | Measure-Object -Maximum).Maximum
}

function Get-ServiceCommandLine { (Get-CimInstance Win32_Service -Filter "Name='$serviceName'").PathName }

Assert-Administrator
if (Get-Service $serviceName -ErrorAction SilentlyContinue) { throw "the $serviceName service already exists; run this on a machine without Databasus" }
if (Test-Path (Join-Path $DataDir 'pgdata')) { throw "$DataDir already holds Databasus data; pass another -DataDir or remove it" }
if (Test-Path $WorkDir) { Remove-Item -Recurse -Force $WorkDir }
New-Item -ItemType Directory -Force $WorkDir | Out-Null

try {
    Write-Host "== silent install of the previous release: $previousInstallerPath"
    $exitCode = Invoke-Setup $previousInstallerPath $previousSetupLog @(
        "/DIR=`"$InstallDir`"", "/DATADIR=`"$DataDir`"", "/BACKUPSDIR=`"$BackupsDir`"", "/PORT=$Port"
    )
    if ($exitCode -ne 0) { throw "the previous release's setup exited with $exitCode" }
    Wait-Until { Test-Healthy } 120 'the API to answer'
    $previousVersion = Get-AppVersion
    Write-Host "previous version: $previousVersion"

    Write-Host '== source database (from a copy of the previous PostgreSQL bundle)'
    $bundleCopy = Join-Path $WorkDir 'postgresql'
    Copy-Item -Recurse (Get-InstalledPostgresqlBundle) $bundleCopy
    $postgresqlBin = Join-Path $bundleCopy 'bin'
    Start-SourcePostgres
    Invoke-Psql 'postgres' 'CREATE DATABASE upgrade_src' | Out-Null
    Invoke-Psql 'postgres' 'CREATE DATABASE upgrade_restored' | Out-Null
    Invoke-Psql 'upgrade_src' @'
CREATE TABLE items (id integer PRIMARY KEY, label text NOT NULL, amount numeric(12, 2) NOT NULL);
INSERT INTO items SELECT i, 'item-' || i, i * 1.25 FROM generate_series(1, 5000) AS i;
'@ | Out-Null
    $sourceChecksum = Get-DataChecksum 'upgrade_src'

    Write-Host '== user, workspace, local storage, database and a backup with the previous release'
    $token = (Invoke-Api POST '/users/signup' @{ email = $userEmail; password = $userPassword; name = 'Upgrade' }).token
    if (-not $token) { throw 'sign-up returned no token' }
    $workspace = Invoke-Api POST '/workspaces' @{ name = 'Upgrade' } $token
    $storage = Invoke-Api POST '/storages' @{ workspaceId = $workspace.id; type = 'LOCAL'; name = 'Backups folder'; localStorage = @{} } $token
    $connection = @{
        host = '127.0.0.1'; port = $SourcePostgresPort; username = 'postgres'; password = $sourcePassword
        database = 'upgrade_src'; sslMode = 'disable'; cpuCount = 1
    }
    $database = Invoke-Api POST '/databases/create' @{
        workspaceId = $workspace.id; name = 'upgrade-src'; type = 'POSTGRES_LOGICAL'; postgresqlLogical = $connection
    } $token

    $backupConfig = Invoke-Api GET "/backup-configs/database/$($database.id)" $null $token
    $backupConfig.isBackupsEnabled = $false
    $backupConfig | Add-Member -NotePropertyName storage -NotePropertyValue @{ id = $storage.id } -Force
    Invoke-Api POST '/backup-configs/save' $backupConfig $token | Out-Null

    $previousBackup = Wait-ForBackup $database.id $token
    $previousBackupFile = Join-Path $BackupsDir $previousBackup.fileName
    if (-not (Test-Path $previousBackupFile)) { throw "the backup file is not in the backups folder: $previousBackupFile" }

    # The schedule is switched on after the manual backup, so a scheduled run cannot be mistaken for it.
    $backupConfig = Invoke-Api GET "/backup-configs/database/$($database.id)" $null $token
    $backupConfig.isBackupsEnabled = $true
    $backupConfig.backupInterval = @{ type = 'WEEKLY'; timeOfDay = '03:30'; weekday = 3 }
    $backupConfig | Add-Member -NotePropertyName storage -NotePropertyValue @{ id = $storage.id } -Force
    Invoke-Api POST '/backup-configs/save' $backupConfig $token | Out-Null
    $scheduleBefore = Get-ScheduleSummary (Invoke-Api GET "/backup-configs/database/$($database.id)" $null $token)
    Write-Host "schedule: $scheduleBefore"
    $migrationBefore = Get-AppliedMigrationVersion
    $commandLineBefore = Get-ServiceCommandLine

    Write-Host '== Setup refuses a metadata database of another PostgreSQL major and leaves the service running'
    $pgVersionFile = Join-Path $DataDir 'pgdata\PG_VERSION'
    $realMajor = (Get-Content $pgVersionFile -Raw).Trim()
    Set-Content -Path $pgVersionFile -Value '17' -NoNewline -Encoding ascii
    try {
        $exitCode = Invoke-Setup $installerPath $refusedSetupLog
    } finally {
        Set-Content -Path $pgVersionFile -Value "$realMajor`n" -NoNewline -Encoding ascii
    }
    if ($exitCode -ne 7) { throw "setup with a PostgreSQL 17 metadata database exited with $exitCode, expected 7" }
    if (-not (Select-String -Path $refusedSetupLog -Pattern 'pg_upgrade' -Quiet)) { throw 'the setup log does not explain the refusal' }
    if ((Get-Service $serviceName).Status -ne 'Running') { throw 'the refused setup stopped the service' }
    if ((Get-AppVersion) -ne $previousVersion) { throw 'the refused setup changed the running version' }

    Write-Host "== silent upgrade with no folder or port options: $installerPath"
    $exitCode = Invoke-Setup $installerPath $upgradeSetupLog
    if ($exitCode -ne 0) { throw "the upgrade setup exited with $exitCode" }
    Wait-Until { Test-Healthy } 240 'the API to answer after the upgrade'
    $currentVersion = Get-AppVersion
    Write-Host "upgraded: $previousVersion -> $currentVersion"
    if ($currentVersion -eq $previousVersion) { throw "the version is still $previousVersion after the upgrade" }
    $commandLine = Get-ServiceCommandLine
    if ($commandLine -ne $commandLineBefore) { Write-Host "service command line was: $commandLineBefore" }
    if (-not $commandLine.Contains($DataDir) -or -not $commandLine.Contains("-port $Port")) {
        throw "the upgraded service does not use the previous data folder and port: $commandLine"
    }
    $recordedBackupsFolder = (Get-Content (Join-Path $DataDir 'backups-folder.txt') | Where-Object { $_ -like 'current=*' }) -replace '^current=', ''
    if ($recordedBackupsFolder -ne $BackupsDir) {
        throw "the upgraded service records the backups folder $recordedBackupsFolder, not $BackupsDir"
    }

    Write-Host '== newest migration applied'
    $migrationAfter = Get-AppliedMigrationVersion
    $newestShipped = Get-NewestShippedMigrationVersion
    Write-Host "migration version: $migrationBefore -> $migrationAfter (newest shipped $newestShipped)"
    if ($migrationAfter -ne $newestShipped) { throw "the metadata database is at migration $migrationAfter, the new version ships $newestShipped" }

    Write-Host '== user, workspace, database, schedule and backup are still there'
    $token = (Invoke-Api POST '/users/signin' @{ email = $userEmail; password = $userPassword }).token
    if (-not $token) { throw 'the user cannot sign in after the upgrade' }
    $workspaces = @((Invoke-Api GET '/workspaces' $null $token).workspaces)
    if (-not ($workspaces | Where-Object id -eq $workspace.id)) { throw 'the workspace is gone after the upgrade' }
    $databaseAfter = Invoke-Api GET "/databases/$($database.id)" $null $token
    if ($databaseAfter.id -ne $database.id) { throw 'the database is gone after the upgrade' }
    $scheduleAfter = Get-ScheduleSummary (Invoke-Api GET "/backup-configs/database/$($database.id)" $null $token)
    if ($scheduleAfter -ne $scheduleBefore) { throw "the schedule changed: before '$scheduleBefore', after '$scheduleAfter'" }
    if (-not (Get-BackupById $database.id $previousBackup.id $token)) { throw 'the backup is gone after the upgrade' }
    if (-not (Test-Path $previousBackupFile)) { throw 'the backup file is gone after the upgrade' }

    Write-Host '== restore the previous release''s backup'
    $restoreTarget = $connection.Clone()
    $restoreTarget.database = 'upgrade_restored'
    Invoke-Api POST "/restores/$($previousBackup.id)/restore" @{ postgresqlDatabase = $restoreTarget } $token | Out-Null
    $restoreStatus = Wait-ForStatus {
        # Invoke-RestMethod returns a JSON array as one object; the pipeline unrolls it.
        $restores = @(Invoke-Api GET "/restores/$($previousBackup.id)" $null $token | ForEach-Object { $_ })
        $statuses = @($restores | Where-Object { $null -ne $_ -and $_.PSObject.Properties['status'] } | ForEach-Object { $_.status })
        if ($statuses -contains 'COMPLETED') { 'COMPLETED' } elseif ($statuses -contains 'FAILED') { 'FAILED' } else { '' }
    } 'the restore to finish'
    if ($restoreStatus -ne 'COMPLETED') { throw "the restore ended $restoreStatus" }
    $restoredChecksum = Get-DataChecksum 'upgrade_restored'
    if ($restoredChecksum -ne $sourceChecksum) { throw "restored data differs: source $sourceChecksum, restored $restoredChecksum" }

    Write-Host '== back up with the new version'
    $newBackup = Wait-ForBackup $database.id $token
    if (-not (Test-Path (Join-Path $BackupsDir $newBackup.fileName))) { throw 'the new backup file is not in the backups folder' }

    Write-Host 'installer upgrade smoke test OK'
} catch {
    Show-Diagnostics
    throw
} finally {
    Stop-SourcePostgres
    if (Test-Path (Join-Path $InstallDir 'unins000.exe')) {
        Start-Process -FilePath (Join-Path $InstallDir 'unins000.exe') -Wait -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES'
    }
}
