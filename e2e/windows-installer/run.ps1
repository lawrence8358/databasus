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
$isVhdAttached = $false
$isSourcePostgresStarted = $false

function Assert-Administrator {
    $principal = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
    if (-not $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
        throw 'run this from an elevated PowerShell (administrator)'
    }
}

function Wait-Until([scriptblock]$Condition, [int]$TimeoutSeconds, [string]$Description) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        if (& $Condition) { return }
        Start-Sleep -Seconds 1
    }
    throw "timed out after $TimeoutSeconds s waiting for: $Description"
}

function Invoke-Diskpart([string[]]$Commands) {
    $scriptPath = Join-Path $WorkDir 'diskpart.txt'
    Set-Content -Path $scriptPath -Value $Commands -Encoding ascii
    $output = & diskpart.exe /s $scriptPath
    if ($LASTEXITCODE -ne 0) { throw "diskpart failed:`n$($output -join "`n")" }
}

# windows-latest has only C:, and local storage must live on another disk to prove the backups
# folder can sit anywhere. A VHD mounts as a real fixed disk that the LocalSystem service can see
# (a subst drive would exist only in this logon session).
function New-SecondDisk {
    $usedLetters = (Get-PSDrive -PSProvider FileSystem).Name
    $letter = [char[]]'RSTUVWXYZ' | Where-Object { $usedLetters -notcontains [string]$_ } | Select-Object -First 1
    if (-not $letter) { throw 'no free drive letter for the second disk' }

    Invoke-Diskpart @(
        "create vdisk file=`"$vhdPath`" maximum=2048 type=expandable",
        "select vdisk file=`"$vhdPath`"",
        'attach vdisk',
        'create partition primary',
        'format fs=ntfs quick label=DbsSmoke',
        "assign letter=$letter"
    )
    $script:isVhdAttached = $true
    Wait-Until { Test-Path "${letter}:\" } 30 "drive ${letter}: to appear"

    return "${letter}:\databasus-backups"
}

function Remove-SecondDisk {
    if (-not $script:isVhdAttached) { return }
    Invoke-Diskpart @("select vdisk file=`"$vhdPath`"", 'detach vdisk')
    Remove-Item $vhdPath -Force -ErrorAction SilentlyContinue
}

function Get-PostgresqlBin { Join-Path $InstallDir 'assets\tools\win-x64\postgresql\postgresql-18\bin' }

function Invoke-Psql([string]$Database, [string]$Sql) {
    $env:PGPASSWORD = $sourcePassword
    $output = & (Join-Path (Get-PostgresqlBin) 'psql.exe') -h 127.0.0.1 -p $SourcePostgresPort -U postgres -d $Database `
        -v ON_ERROR_STOP=1 -X -A -t -c $Sql
    if ($LASTEXITCODE -ne 0) { throw "psql failed on ${Database}: $Sql" }
    return ($output -join "`n").Trim()
}

# Uses the PostgreSQL the installer just laid down, so the test needs nothing else installed and
# also proves the bundled binaries run (VC++ runtime included) on this machine.
function Start-SourcePostgres {
    $bin = Get-PostgresqlBin
    $passwordFile = Join-Path $WorkDir 'source-password.txt'
    Set-Content -Path $passwordFile -Value $sourcePassword -NoNewline -Encoding ascii

    & (Join-Path $bin 'initdb.exe') -D $sourceDataDir -U postgres -A scram-sha-256 "--pwfile=$passwordFile" -E UTF8 --locale=C | Out-Host
    if ($LASTEXITCODE -ne 0) { throw 'initdb of the source server failed' }

    # The server inherits pg_ctl's output handles. Files keep them away from this script's own
    # output pipe, which would otherwise stay open (and block the caller) while the server runs.
    $pgCtlLog = Join-Path $WorkDir 'source-pg_ctl.log'
    $pgCtl = Start-Process -FilePath (Join-Path $bin 'pg_ctl.exe') -NoNewWindow -PassThru `
        -RedirectStandardOutput $pgCtlLog -RedirectStandardError "$pgCtlLog.err" -ArgumentList @(
        '-D', "`"$sourceDataDir`"", '-l', "`"$(Join-Path $WorkDir 'source-postgres.log')`"", '-w',
        '-o', "`"-p $SourcePostgresPort -c listen_addresses=127.0.0.1`"", 'start'
    )
    $null = $pgCtl.Handle
    $pgCtl.WaitForExit()
    if ($pgCtl.ExitCode -ne 0) {
        Get-Content $pgCtlLog, "$pgCtlLog.err" -ErrorAction SilentlyContinue | Out-Host
        throw 'the source PostgreSQL server did not start'
    }
    $script:isSourcePostgresStarted = $true
}

function Stop-SourcePostgres {
    if (-not $script:isSourcePostgresStarted) { return }
    & (Join-Path (Get-PostgresqlBin) 'pg_ctl.exe') -D $sourceDataDir -m fast -w stop | Out-Host
}

function Invoke-Api([string]$Method, [string]$Path, $Body = $null, [string]$Token = '') {
    $parameters = @{ Method = $Method; Uri = "$baseUrl$Path"; ContentType = 'application/json'; TimeoutSec = 30 }
    if ($null -ne $Body) { $parameters.Body = ($Body | ConvertTo-Json -Depth 10) }
    if ($Token) { $parameters.Headers = @{ Authorization = "Bearer $Token" } }
    Invoke-RestMethod @parameters
}

function Test-Healthy {
    try { (Invoke-WebRequest -Uri "$baseUrl/system/version" -UseBasicParsing -TimeoutSec 3).StatusCode -eq 200 } catch { $false }
}

function Get-DataChecksum([string]$Database) {
    Invoke-Psql $Database "SELECT count(*) || ':' || md5(string_agg(id || '|' || label || '|' || amount, ',' ORDER BY id)) FROM items"
}

function Wait-ForStatus([scriptblock]$GetStatus, [string]$Description) {
    Wait-Until {
        $script:lastStatus = & $GetStatus
        $script:lastStatus -in 'COMPLETED', 'FAILED', 'CANCELED'
    } 300 $Description
    return $script:lastStatus
}

function Show-Diagnostics {
    Write-Host '--- setup log (tail) ---'
    if (Test-Path $setupLog) { Get-Content $setupLog -Tail 60 }
    Write-Host '--- service ---'
    Get-CimInstance Win32_Service -Filter "Name='$serviceName'" | Format-List Name, State, StartMode, ProcessId, ExitCode, PathName
    $launcherLog = Join-Path $DataDir 'logs\launcher.log'
    Write-Host '--- launcher.log (tail) ---'
    if (Test-Path $launcherLog) { Get-Content $launcherLog -Tail 100 }
    Write-Host '--- events ---'
    try {
        Get-WinEvent -FilterHashtable @{ LogName = 'Application'; ProviderName = $serviceName } -MaxEvents 10 |
            Format-List TimeCreated, LevelDisplayName, Message
    } catch {
        Write-Host "no events: $_"
    }
    Write-Host '--- uninstall log (tail) ---'
    if (Test-Path $uninstallLog) { Get-Content $uninstallLog -Tail 40 }
}

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
