# Drives the Databasus Windows service through the service control manager: install, start,
# restart, a forced kill that the recovery actions must undo, an orderly stop, a start failure
# that must reach the event log, and uninstall. Needs administrator rights and databasus.exe built
# beside the launcher inside a checkout (the launcher finds the bundled PostgreSQL from there).
param(
    [Parameter(Mandatory = $true)][string]$LauncherPath,
    [Parameter(Mandatory = $true)][string]$DataDir
)

$ErrorActionPreference = 'Stop'

$serviceName = 'Databasus'
$healthUrl = 'http://127.0.0.1:4005/api/v1/system/version'
$launcherPath = (Resolve-Path $LauncherPath).Path
$installFolder = Split-Path $launcherPath
$launcherLog = Join-Path $DataDir 'logs\launcher.log'

function Invoke-Launcher([string[]]$Arguments) {
    & $launcherPath @Arguments
    if ($LASTEXITCODE -ne 0) { throw "databasus-launcher $($Arguments -join ' ') exited with $LASTEXITCODE" }
}

function Get-ServiceInfo {
    Get-CimInstance Win32_Service -Filter "Name='$serviceName'"
}

function Wait-Until([scriptblock]$Condition, [int]$TimeoutSeconds, [string]$Description) {
    $deadline = (Get-Date).AddSeconds($TimeoutSeconds)
    while ((Get-Date) -lt $deadline) {
        if (& $Condition) { return }
        Start-Sleep -Seconds 1
    }
    throw "timed out after $TimeoutSeconds s waiting for: $Description"
}

function Test-Healthy {
    try {
        (Invoke-WebRequest -Uri $healthUrl -UseBasicParsing -TimeoutSec 3).StatusCode -eq 200
    } catch {
        $false
    }
}

function Wait-Healthy {
    Wait-Until { Test-Healthy } 240 'the backend to answer on its health endpoint'
}

function Get-ChildProcesses {
    Get-CimInstance Win32_Process |
        Where-Object {
            ($_.Name -eq 'databasus.exe' -and $_.ExecutablePath -like "$installFolder\*") -or
            ($_.Name -eq 'postgres.exe' -and $_.ExecutablePath -like '*\assets\tools\win-x64\*')
        }
}

function Get-ServiceEvents {
    @(Get-WinEvent -FilterHashtable @{ LogName = 'Application'; ProviderName = $serviceName } -ErrorAction SilentlyContinue)
}

function Get-GracefulStopCount {
    if (-not (Test-Path $launcherLog)) { return 0 }
    @(Select-String -Path $launcherLog -SimpleMatch 'server gracefully stopped').Count
}

function Show-Diagnostics {
    Write-Host '--- service ---'
    Get-ServiceInfo | Format-List Name, State, StartMode, StartName, ProcessId, ExitCode, PathName
    Write-Host '--- events ---'
    Get-ServiceEvents | Select-Object -First 20 | Format-List TimeCreated, LevelDisplayName, Message
    Write-Host '--- launcher.log (tail) ---'
    if (Test-Path $launcherLog) { Get-Content $launcherLog -Tail 80 }
}

try {
    Write-Host '== install'
    Invoke-Launcher @('install', '-data-dir', $DataDir)

    $service = Get-ServiceInfo
    if ($service.StartMode -ne 'Auto') { throw "start mode is $($service.StartMode), expected Auto" }
    if ($service.StartName -ne 'LocalSystem') { throw "account is $($service.StartName), expected LocalSystem" }

    $failureActions = (sc.exe qfailure $serviceName) -join "`n"
    if ($failureActions -notmatch 'RESTART') { throw "no restart recovery action:`n$failureActions" }
    $nonCrashFlag = (sc.exe qfailureflag $serviceName) -join "`n"
    if ($nonCrashFlag -notmatch 'TRUE') { throw "recovery does not cover error exits:`n$nonCrashFlag" }

    Write-Host '== start'
    Start-Service $serviceName
    Wait-Healthy
    $passwordFile = Join-Path $DataDir 'metadata-db.password'
    $passwordHash = (Get-FileHash $passwordFile).Hash

    Write-Host '== restart'
    $gracefulStopsBefore = Get-GracefulStopCount
    Restart-Service $serviceName
    Wait-Healthy
    if ((Get-GracefulStopCount) -le $gracefulStopsBefore) { throw 'the backend was not stopped gracefully on restart' }

    Write-Host '== forced kill is undone by the recovery actions'
    $killedProcessId = (Get-ServiceInfo).ProcessId
    Stop-Process -Id $killedProcessId -Force
    Wait-Until { @(Get-ChildProcesses).Count -eq 0 } 8 'the backend and PostgreSQL to die with the killed launcher'
    Wait-Until {
        $current = Get-ServiceInfo
        $current.State -eq 'Running' -and $current.ProcessId -ne 0 -and $current.ProcessId -ne $killedProcessId
    } 90 'the service control manager to restart the service'
    Wait-Healthy
    if ((Get-FileHash $passwordFile).Hash -ne $passwordHash) { throw 'the metadata database was initialised again instead of reused' }

    Write-Host '== stop'
    $gracefulStopsBefore = Get-GracefulStopCount
    Stop-Service $serviceName
    if ((Get-ServiceInfo).State -ne 'Stopped') { throw 'the service did not stop' }
    if (@(Get-ChildProcesses).Count -ne 0) { throw 'the backend or PostgreSQL outlived the stopped service' }
    if ((Get-GracefulStopCount) -le $gracefulStopsBefore) { throw 'the backend was not stopped gracefully' }
    if (Select-String -Path $launcherLog -SimpleMatch 'did not stop in time') { throw 'a child had to be killed on stop' }

    $messages = (Get-ServiceEvents).Message -join "`n"
    if ($messages -notmatch 'Databasus service started' -or $messages -notmatch 'Databasus service stopped') {
        throw "start and stop are missing from the event log:`n$messages"
    }

    Write-Host '== a start failure reaches the event log'
    Invoke-Launcher @('install', '-data-dir', $DataDir, '-backend', (Join-Path $installFolder 'missing\databasus.exe'))
    Start-Service $serviceName
    Wait-Until {
        ((Get-ServiceEvents | Where-Object LevelDisplayName -eq 'Error').Message -join "`n") -match 'backend executable'
    } 60 'the start failure in the event log'

    Write-Host '== uninstall'
    Invoke-Launcher @('uninstall')
    if (Get-Service $serviceName -ErrorAction SilentlyContinue) { throw 'the service is still registered' }
    if (-not (Test-Path (Join-Path $DataDir 'pgdata'))) { throw 'uninstall removed the data folder' }
    if (@(Get-ChildProcesses).Count -ne 0) { throw 'processes outlived the uninstalled service' }

    Write-Host 'service lifecycle OK'
} catch {
    Show-Diagnostics
    if (Get-Service $serviceName -ErrorAction SilentlyContinue) { & $launcherPath uninstall }
    throw
}
