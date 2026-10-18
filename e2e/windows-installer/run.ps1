# End-to-end smoke test of the offline Windows installer, driven only through its public surfaces:
# the installer and uninstaller command lines, the Windows service, and the HTTP API.
#
#   silent install -> service running, API answers on 127.0.0.1 only -> create a user, a workspace,
#   a local storage (the install-time backups folder on a second disk) and a local PostgreSQL
#   database -> back up -> restore into another database and compare the data -> restart the
#   service and check the state is kept -> silent uninstall keeps the data and backups folders.
#
# Needs an elevated PowerShell 7 on a machine without Databasus installed. It creates a 2 GB VHD as
# the second disk unless -BackupsDir is given, and runs a throwaway PostgreSQL server (from the
# installed bundle) on -SourcePostgresPort as the database to back up. How to run it by hand is in
# packaging/windows/README.md.
param(
    [Parameter(Mandatory = $true)][string]$InstallerPath,
    [string]$InstallDir = (Join-Path $env:ProgramFiles 'Databasus'),
    [string]$DataDir = (Join-Path $env:ProgramData 'Databasus'),
    [string]$BackupsDir = '',
    [int]$Port = 4123,
    [int]$SourcePostgresPort = 5499,
    [string]$WorkDir = (Join-Path ([System.IO.Path]::GetTempPath()) 'databasus-installer-smoke')
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$serviceName = 'Databasus'
$baseUrl = "http://127.0.0.1:$Port/api/v1"
$installerPath = (Resolve-Path $InstallerPath).Path
$sourcePassword = 'smoke-source-password'
$userEmail = 'smoke@example.com'
$userPassword = 'Smoke-Passw0rd!'
$setupLog = Join-Path $WorkDir 'setup.log'
$uninstallLog = Join-Path $WorkDir 'uninstall.log'
$sourceDataDir = Join-Path $WorkDir 'source-pgdata'
$vhdPath = Join-Path $WorkDir 'second-disk.vhdx'

. (Join-Path $PSScriptRoot 'common.ps1')

Assert-Administrator
if (Get-Service $serviceName -ErrorAction SilentlyContinue) { throw "the $serviceName service already exists; run this on a machine without Databasus" }
if (Test-Path (Join-Path $DataDir 'pgdata')) { throw "$DataDir already holds Databasus data; pass another -DataDir or remove it" }
if (Test-Path $WorkDir) { Remove-Item -Recurse -Force $WorkDir }
New-Item -ItemType Directory -Force $WorkDir | Out-Null

try {
    if (-not $BackupsDir) {
        Write-Host '== second disk'
        $BackupsDir = New-SecondDisk
    }
    Write-Host "backups folder: $BackupsDir"

    Write-Host '== silent install'
    $setup = Start-Process -FilePath $installerPath -Wait -PassThru -ArgumentList @(
        '/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', "/LOG=`"$setupLog`"",
        "/DIR=`"$InstallDir`"", "/DATADIR=`"$DataDir`"", "/BACKUPSDIR=`"$BackupsDir`"", "/PORT=$Port"
    )
    if ($setup.ExitCode -ne 0) { throw "setup exited with $($setup.ExitCode)" }

    Write-Host '== service running and API answering on 127.0.0.1 only'
    $service = Get-CimInstance Win32_Service -Filter "Name='$serviceName'"
    if ($service.State -ne 'Running') { throw "the service is $($service.State)" }
    if ($service.StartMode -ne 'Auto') { throw "the service start mode is $($service.StartMode)" }
    Wait-Until { Test-Healthy } 120 'the API to answer'
    $listeners = @(Get-NetTCPConnection -LocalPort $Port -State Listen)
    $foreignListeners = @($listeners | Where-Object LocalAddress -ne '127.0.0.1')
    if ($listeners.Count -eq 0 -or $foreignListeners.Count -gt 0) {
        throw "the Web UI listens on: $(($listeners.LocalAddress) -join ', '); expected 127.0.0.1 only"
    }

    Write-Host '== source database'
    Start-SourcePostgres
    Invoke-Psql 'postgres' 'CREATE DATABASE smoke_src' | Out-Null
    Invoke-Psql 'postgres' 'CREATE DATABASE smoke_restored' | Out-Null
    Invoke-Psql 'smoke_src' @'
CREATE TABLE items (id integer PRIMARY KEY, label text NOT NULL, amount numeric(12, 2) NOT NULL);
INSERT INTO items SELECT i, 'item-' || i, i * 1.25 FROM generate_series(1, 5000) AS i;
'@ | Out-Null
    $sourceChecksum = Get-DataChecksum 'smoke_src'

    Write-Host '== user, workspace, local storage, database'
    $token = (Invoke-Api POST '/users/signup' @{ email = $userEmail; password = $userPassword; name = 'Smoke' }).token
    if (-not $token) { throw 'sign-up returned no token' }
    $workspace = Invoke-Api POST '/workspaces' @{ name = 'Smoke' } $token
    $storage = Invoke-Api POST '/storages' @{ workspaceId = $workspace.id; type = 'LOCAL'; name = 'Second disk'; localStorage = @{} } $token
    $connection = @{
        host = '127.0.0.1'; port = $SourcePostgresPort; username = 'postgres'; password = $sourcePassword
        database = 'smoke_src'; sslMode = 'disable'; cpuCount = 1
    }
    $database = Invoke-Api POST '/databases/create' @{
        workspaceId = $workspace.id; name = 'smoke-src'; type = 'POSTGRES_LOGICAL'; postgresqlLogical = $connection
    } $token

    $backupConfig = Invoke-Api GET "/backup-configs/database/$($database.id)" $null $token
    $backupConfig.isBackupsEnabled = $false
    $backupConfig | Add-Member -NotePropertyName storage -NotePropertyValue @{ id = $storage.id } -Force
    Invoke-Api POST '/backup-configs/save' $backupConfig $token | Out-Null

    Write-Host '== backup'
    Invoke-Api POST '/backups' @{ database_id = $database.id } $token | Out-Null
    $backupStatus = Wait-ForStatus {
        $latest = @((Invoke-Api GET "/backups?database_id=$($database.id)&limit=1" $null $token).backups)
        if ($latest.Count -gt 0) { $latest[0].status } else { '' }
    } 'the backup to finish'
    $backup = @((Invoke-Api GET "/backups?database_id=$($database.id)&limit=1" $null $token).backups)[0]
    if ($backupStatus -ne 'COMPLETED') { throw "the backup ended $backupStatus`: $($backup.failMessage)" }
    $backupFile = Join-Path $BackupsDir $backup.fileName
    if (-not (Test-Path $backupFile)) { throw "the backup file is not in the backups folder: $backupFile" }
    Write-Host "backup file: $backupFile"

    Write-Host '== restore into another database'
    $restoreTarget = $connection.Clone()
    $restoreTarget.database = 'smoke_restored'
    Invoke-Api POST "/restores/$($backup.id)/restore" @{ postgresqlDatabase = $restoreTarget } $token | Out-Null
    $restoreStatus = Wait-ForStatus {
        # Invoke-RestMethod returns a JSON array as one object; the pipeline unrolls it.
        $restores = @(Invoke-Api GET "/restores/$($backup.id)" $null $token | ForEach-Object { $_ })
        $statuses = @($restores | Where-Object { $null -ne $_ -and $_.PSObject.Properties['status'] } | ForEach-Object { $_.status })
        if ($statuses -contains 'COMPLETED') { 'COMPLETED' } elseif ($statuses -contains 'FAILED') { 'FAILED' } else { '' }
    } 'the restore to finish'
    if ($restoreStatus -ne 'COMPLETED') { throw "the restore ended $restoreStatus" }
    $restoredChecksum = Get-DataChecksum 'smoke_restored'
    if ($restoredChecksum -ne $sourceChecksum) { throw "restored data differs: source $sourceChecksum, restored $restoredChecksum" }

    Write-Host '== restart keeps the state'
    Restart-Service $serviceName
    Wait-Until { Test-Healthy } 240 'the API to answer after the restart'
    $token = (Invoke-Api POST '/users/signin' @{ email = $userEmail; password = $userPassword }).token
    $workspaces = @((Invoke-Api GET '/workspaces' $null $token).workspaces)
    if (-not ($workspaces | Where-Object id -eq $workspace.id)) { throw 'the workspace is gone after the restart' }
    $backupsAfterRestart = @((Invoke-Api GET "/backups?database_id=$($database.id)" $null $token).backups)
    if (-not ($backupsAfterRestart | Where-Object id -eq $backup.id)) { throw 'the backup is gone after the restart' }

    Write-Host '== silent uninstall keeps the data'
    Stop-SourcePostgres
    $script:isSourcePostgresStarted = $false
    $uninstall = Start-Process -FilePath (Join-Path $InstallDir 'unins000.exe') -Wait -PassThru `
        -ArgumentList @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', "/LOG=`"$uninstallLog`"")
    if ($uninstall.ExitCode -ne 0) { throw "the uninstaller exited with $($uninstall.ExitCode)" }
    Wait-Until { -not (Test-Path (Join-Path $InstallDir 'databasus.exe')) } 120 'the program files to be removed'
    if (Get-Service $serviceName -ErrorAction SilentlyContinue) { throw 'the service is still registered' }
    if (-not (Test-Path (Join-Path $DataDir 'pgdata'))) { throw 'the uninstaller removed the program data folder' }
    if (-not (Test-Path $backupFile)) { throw 'the uninstaller removed the backups' }

    Write-Host 'installer smoke test OK'
} catch {
    Show-Diagnostics
    throw
} finally {
    Stop-SourcePostgres
    if (Test-Path (Join-Path $InstallDir 'unins000.exe')) {
        Start-Process -FilePath (Join-Path $InstallDir 'unins000.exe') -Wait -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES'
    }
    Remove-SecondDisk
}
