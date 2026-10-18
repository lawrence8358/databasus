# Helpers shared by the installer smoke tests (run.ps1, upgrade.ps1). Dot-source this file after the
# caller has set the variables these functions read:
#   $WorkDir, $InstallDir, $DataDir, $serviceName, $baseUrl, $setupLog, $uninstallLog, $vhdPath,
#   $sourcePassword, $SourcePostgresPort, $sourceDataDir
# and optionally $postgresqlBin, the bin folder of a PostgreSQL bundle to run the source server from
# (default: the bundle in $InstallDir).

$isVhdAttached = $false
$isSourcePostgresStarted = $false
$postgresqlBin = ''

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

function Get-InstalledPostgresqlBundle { Join-Path $InstallDir 'assets\tools\win-x64\postgresql\postgresql-18' }

function Get-PostgresqlBin {
    if ($postgresqlBin) { return $postgresqlBin }
    Join-Path (Get-InstalledPostgresqlBundle) 'bin'
}

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
