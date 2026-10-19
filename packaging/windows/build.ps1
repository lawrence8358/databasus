# Builds the offline Windows installer: compiles the backend and the launcher, stages them with the
# migrations, the built Web UI, the bundled PostgreSQL and the VC++ runtime, and runs Inno Setup.
# Run from any folder: pwsh packaging/windows/build.ps1 -Version 1.2.3
# Needs Go, swag (for the swagger docs the backend embeds), a built frontend (frontend/dist) and
# Inno Setup 7 (ISCC.exe on PATH, in its default install folder, or passed with -IsccPath).
param(
    [string]$Version = '0.0.0-dev',
    [string]$IsccPath = '',
    [string]$FrontendDist = '',
    [string]$OutputDir = ''
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
$backendFolder = Join-Path $repoRoot 'backend'
if (-not $FrontendDist) { $FrontendDist = Join-Path $repoRoot 'frontend\dist' }
if (-not $OutputDir) { $OutputDir = Join-Path $PSScriptRoot 'out' }

$stageFolder = Join-Path $OutputDir 'stage'
$postgresqlBundle = 'assets\tools\win-x64\postgresql\postgresql-18'

# The bundled PostgreSQL binaries import these and the EDB archive does not ship them. Windows 10
# and Server 2016 include the Universal CRT, so these three are all a clean machine lacks.
$vcRuntimeDlls = @('vcruntime140.dll', 'vcruntime140_1.dll', 'msvcp140.dll')

function Find-Iscc {
    if ($IsccPath) { return (Resolve-Path $IsccPath).Path }

    $onPath = Get-Command 'ISCC.exe' -ErrorAction SilentlyContinue
    if ($onPath) { return $onPath.Source }

    foreach ($programFiles in @(${env:ProgramFiles(x86)}, $env:ProgramFiles, "$env:LOCALAPPDATA\Programs")) {
        if (-not $programFiles) { continue }
        $candidate = Join-Path $programFiles 'Inno Setup 7\ISCC.exe'
        if (Test-Path $candidate) { return $candidate }
    }

    throw 'ISCC.exe (Inno Setup 7) not found; install it or pass -IsccPath'
}

# Prefers the redistributable folder of an installed Visual Studio, which is the copy Microsoft
# intends for app-local deployment; the copies the VC++ redistributable put in System32 are the
# same files.
function Find-VcRuntimeFolder {
    $vswhere = Join-Path ${env:ProgramFiles(x86)} 'Microsoft Visual Studio\Installer\vswhere.exe'
    if (Test-Path $vswhere) {
        $found = & $vswhere -latest -products * -find 'VC\Redist\MSVC\*\x64\Microsoft.VC*.CRT\vcruntime140.dll' |
            Sort-Object | Select-Object -Last 1
        if ($found) { return Split-Path $found }
    }

    $system32 = Join-Path $env:SystemRoot 'System32'
    if (Test-Path (Join-Path $system32 'vcruntime140_1.dll')) { return $system32 }

    throw 'the VC++ runtime DLLs were not found; install Visual Studio C++ build tools or the VC++ 2015-2022 x64 redistributable'
}

function Invoke-Checked([string]$FilePath, [string[]]$Arguments) {
    & $FilePath @Arguments
    if ($LASTEXITCODE -ne 0) { throw "$FilePath exited with $LASTEXITCODE" }
}

if (-not (Test-Path (Join-Path $FrontendDist 'index.html'))) {
    throw "no built Web UI in $FrontendDist; run 'pnpm build' in frontend first"
}
if (-not (Test-Path (Join-Path $backendFolder 'swagger\docs.go'))) {
    throw "no swagger docs in backend\swagger; run 'swag init -d . -g cmd/main.go -o swagger' in backend first"
}

$iscc = Find-Iscc
$vcRuntimeFolder = Find-VcRuntimeFolder

Write-Host "== staging into $stageFolder"
if (Test-Path $stageFolder) { Remove-Item -Recurse -Force $stageFolder }
New-Item -ItemType Directory -Force $stageFolder | Out-Null

$savedGoEnvironment = @{ CGO_ENABLED = $env:CGO_ENABLED; GOOS = $env:GOOS; GOARCH = $env:GOARCH }
$env:CGO_ENABLED = '0'
$env:GOOS = 'windows'
$env:GOARCH = 'amd64'
Push-Location $backendFolder
try {
    Invoke-Checked 'go' @('build', '-trimpath', '-o', (Join-Path $stageFolder 'databasus.exe'), './cmd')
    Invoke-Checked 'go' @('build', '-trimpath', '-ldflags', "-X main.Version=$Version",
        '-o', (Join-Path $stageFolder 'databasus-launcher.exe'), './cmd/launcher')
} finally {
    Pop-Location
    foreach ($name in $savedGoEnvironment.Keys) {
        [Environment]::SetEnvironmentVariable($name, $savedGoEnvironment[$name])
    }
}

Copy-Item -Recurse (Join-Path $backendFolder 'migrations') (Join-Path $stageFolder 'migrations')
New-Item -ItemType Directory -Force (Join-Path $stageFolder 'ui') | Out-Null
Copy-Item -Recurse $FrontendDist (Join-Path $stageFolder 'ui\build')

$stagedBundle = Join-Path $stageFolder $postgresqlBundle
New-Item -ItemType Directory -Force (Split-Path $stagedBundle) | Out-Null
Copy-Item -Recurse (Join-Path $repoRoot $postgresqlBundle) $stagedBundle

Write-Host "== VC++ runtime from $vcRuntimeFolder"
foreach ($dll in $vcRuntimeDlls) {
    Copy-Item (Join-Path $vcRuntimeFolder $dll) (Join-Path $stagedBundle 'bin')
}

Write-Host "== compiling the installer with $iscc"
Invoke-Checked $iscc @(
    "/Qp",
    "/DAppVersion=$Version",
    "/DStageDir=$stageFolder",
    "/O$OutputDir",
    (Join-Path $PSScriptRoot 'databasus.iss')
)

Get-ChildItem $OutputDir -Filter 'databasus-setup-*.exe' | ForEach-Object { Write-Host "built $($_.FullName)" }
