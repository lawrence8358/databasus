# End-to-end test of the installer's LAN access option, driven through the installer command lines,
# the Windows service, Windows Firewall and HTTPS:
#
#   silent install with /LANACCESS /SELFSIGNED -> the Web UI listens on every interface, answers
#   HTTPS on this computer's network address, refuses plain HTTP, the firewall rule allows the port,
#   the private key is readable only by SYSTEM, Administrators and the installing account -> rerun
#   with /PFX (an imported certificate) -> rerun with no options keeps LAN access and that
#   certificate -> rerun with /LOCALONLY goes back to 127.0.0.1 and removes the firewall rule ->
#   rerun with /LANACCESS, then the silent uninstall removes the firewall rule.
#
# Needs an elevated PowerShell 7 on a machine without Databasus installed. How to run it by hand is
# in packaging/windows/README.md.
param(
    [Parameter(Mandatory = $true)][string]$InstallerPath,
    [string]$InstallDir = (Join-Path $env:ProgramFiles 'Databasus'),
    [string]$DataDir = (Join-Path $env:ProgramData 'Databasus-lan-test'),
    [int]$Port = 4125,
    [string]$WorkDir = (Join-Path ([System.IO.Path]::GetTempPath()) 'databasus-installer-lan')
)

$ErrorActionPreference = 'Stop'
Set-StrictMode -Version Latest

$serviceName = 'Databasus'
$firewallRuleName = 'Databasus Web UI'
$baseUrl = "http://127.0.0.1:$Port/api/v1"
$installerPath = (Resolve-Path $InstallerPath).Path
$setupLog = Join-Path $WorkDir 'setup.log'
$uninstallLog = Join-Path $WorkDir 'uninstall.log'
$sourceDataDir = Join-Path $WorkDir 'unused-pgdata'
$vhdPath = Join-Path $WorkDir 'unused.vhdx'
$pfxPath = Join-Path $WorkDir 'imported.pfx'
$pfxPasswordPath = Join-Path $WorkDir 'imported.pfx.password'
$pfxPassword = 'Lan-Pfx-Passw0rd!'

. (Join-Path $PSScriptRoot 'common.ps1')

function Invoke-Setup([string]$Name, [string[]]$Options = @()) {
    $script:setupLog = Join-Path $WorkDir "setup-$Name.log"
    $arguments = @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', "/LOG=`"$script:setupLog`"",
        "/DIR=`"$InstallDir`"", "/DATADIR=`"$DataDir`"", "/PORT=$Port") + $Options
    $exitCode = (Start-Process -FilePath $installerPath -Wait -PassThru -ArgumentList $arguments).ExitCode
    if ($exitCode -ne 0) { throw "setup ($Name) exited with $exitCode" }
    if (Select-String -Path $script:setupLog -Pattern $pfxPassword -SimpleMatch -Quiet) {
        throw "the setup log ($Name) contains the PFX password"
    }
}

function Get-NetworkAddress {
    $address = Get-NetIPAddress -AddressFamily IPv4 |
        Where-Object { $_.IPAddress -ne '127.0.0.1' -and $_.IPAddress -notlike '169.254.*' } |
        Select-Object -First 1
    if (-not $address) { throw 'this computer has no network address besides loopback' }
    $address.IPAddress
}

# Returns the SHA-256 fingerprint of the certificate the server presents, without trusting it.
function Get-ServedFingerprint([string]$HostName) {
    $client = [System.Net.Sockets.TcpClient]::new($HostName, $Port)
    try {
        $stream = [System.Net.Security.SslStream]::new($client.GetStream(), $false, { $true })
        $stream.AuthenticateAsClient('localhost')
        $certificate = [System.Security.Cryptography.X509Certificates.X509Certificate2]::new($stream.RemoteCertificate)
        $certificate.GetCertHashString([System.Security.Cryptography.HashAlgorithmName]::SHA256)
    } finally {
        $client.Dispose()
    }
}

function Test-HttpsHealthy([string]$HostName) {
    try {
        (Invoke-WebRequest -Uri "https://${HostName}:$Port/api/v1/system/version" -SkipCertificateCheck -TimeoutSec 3).StatusCode -eq 200
    } catch { $false }
}

function Get-FirewallRule { Get-NetFirewallRule -DisplayName $firewallRuleName -ErrorAction SilentlyContinue }

function Assert-LanAccess([string]$Expectation) {
    Wait-Until { Test-HttpsHealthy 'localhost' } 240 "the Web UI to answer over HTTPS ($Expectation)"
    $listeners = @(Get-NetTCPConnection -LocalPort $Port -State Listen)
    # A dual-stack socket shows as :: only.
    $addresses = @($listeners.LocalAddress | Sort-Object -Unique)
    if ($addresses.Count -eq 0 -or @($addresses | Where-Object { $_ -notin '0.0.0.0', '::' }).Count -gt 0) {
        throw "the Web UI listens on $($addresses -join ', '); expected every interface"
    }

    $networkAddress = Get-NetworkAddress
    if (-not (Test-HttpsHealthy $networkAddress)) { throw "no HTTPS answer on $networkAddress" }
    try {
        $plain = Invoke-WebRequest -Uri "http://${networkAddress}:$Port/api/v1/system/version" -TimeoutSec 5 -SkipHttpErrorCheck
        if ($plain.StatusCode -eq 200) { throw 'plain HTTP is served' }
    } catch [System.Net.Http.HttpRequestException] {
        # A refused or reset connection is fine too.
    }

    $rule = Get-FirewallRule
    if (-not $rule) { throw "the firewall rule '$firewallRuleName' is missing" }
    $portFilter = $rule | Get-NetFirewallPortFilter
    $applicationFilter = $rule | Get-NetFirewallApplicationFilter
    if ($rule.Direction -ne 'Inbound' -or $rule.Action -ne 'Allow' -or "$($portFilter.LocalPort)" -ne "$Port") {
        throw "unexpected firewall rule: $($rule.Direction) $($rule.Action) port $($portFilter.LocalPort)"
    }
    if ($applicationFilter.Program -ne (Join-Path $InstallDir 'databasus.exe')) {
        throw "the firewall rule is for $($applicationFilter.Program)"
    }
}

function Assert-KeyIsRestricted {
    $keyPath = Join-Path $DataDir 'tls\databasus.key'
    $acl = Get-Acl $keyPath
    if (-not $acl.AreAccessRulesProtected) { throw "$keyPath inherits permissions" }
    $installer = [Security.Principal.WindowsIdentity]::GetCurrent().User
    $allowed = @('S-1-5-18', 'S-1-5-32-544', $installer.Value)
    foreach ($rule in $acl.GetAccessRules($true, $true, [Security.Principal.SecurityIdentifier])) {
        if ($rule.IdentityReference.Value -notin $allowed) { throw "$keyPath grants access to $($rule.IdentityReference)" }
    }
}

function New-TestPfx {
    $key = [System.Security.Cryptography.RSA]::Create(2048)
    $request = [System.Security.Cryptography.X509Certificates.CertificateRequest]::new(
        'CN=databasus-lan-test.example', $key, [System.Security.Cryptography.HashAlgorithmName]::SHA256,
        [System.Security.Cryptography.RSASignaturePadding]::Pkcs1)
    $names = [System.Security.Cryptography.X509Certificates.SubjectAlternativeNameBuilder]::new()
    $names.AddDnsName('localhost')
    $request.CertificateExtensions.Add($names.Build())
    $certificate = $request.CreateSelfSigned((Get-Date).AddDays(-1), (Get-Date).AddDays(90))
    [System.IO.File]::WriteAllBytes($pfxPath, $certificate.Export('Pfx', $pfxPassword))
    Set-Content -Path $pfxPasswordPath -Value $pfxPassword -Encoding utf8NoBOM
    $certificate.GetCertHashString([System.Security.Cryptography.HashAlgorithmName]::SHA256)
}

Assert-Administrator
if (Get-Service $serviceName -ErrorAction SilentlyContinue) { throw "the $serviceName service already exists; run this on a machine without Databasus" }
if (Get-FirewallRule) { throw "the firewall rule '$firewallRuleName' already exists" }
if (Test-Path (Join-Path $DataDir 'pgdata')) { throw "$DataDir already holds Databasus data; pass another -DataDir or remove it" }
if (Test-Path $WorkDir) { Remove-Item -Recurse -Force $WorkDir }
New-Item -ItemType Directory -Force $WorkDir | Out-Null

try {
    Write-Host '== install with LAN access and a self-signed certificate'
    Invoke-Setup 'self-signed' @('/LANACCESS', '/SELFSIGNED')
    Assert-LanAccess 'self-signed'
    Assert-KeyIsRestricted
    $selfSignedFingerprint = Get-ServedFingerprint (Get-NetworkAddress)
    Write-Host "self-signed certificate: $selfSignedFingerprint"

    Write-Host '== import a PFX certificate'
    $importedFingerprint = New-TestPfx
    Invoke-Setup 'pfx' @('/LANACCESS', "/PFX=`"$pfxPath`"", "/PFXPASSWORDFILE=`"$pfxPasswordPath`"")
    Assert-LanAccess 'imported'
    Assert-KeyIsRestricted
    $served = Get-ServedFingerprint (Get-NetworkAddress)
    if ($served -ne $importedFingerprint) { throw "the server presents $served; expected the imported $importedFingerprint" }

    Write-Host '== rerun without options keeps LAN access and the certificate'
    Invoke-Setup 'keep'
    Assert-LanAccess 'kept'
    $served = Get-ServedFingerprint 'localhost'
    if ($served -ne $importedFingerprint) { throw "the server presents $served after a rerun; expected $importedFingerprint" }

    Write-Host '== /LOCALONLY goes back to 127.0.0.1 and removes the firewall rule'
    Invoke-Setup 'local-only' @('/LOCALONLY')
    Wait-Until { Test-Healthy } 240 'the Web UI to answer over HTTP on 127.0.0.1'
    $addresses = @(Get-NetTCPConnection -LocalPort $Port -State Listen | ForEach-Object LocalAddress | Sort-Object -Unique)
    if (($addresses -join ',') -ne '127.0.0.1') { throw "the Web UI listens on $($addresses -join ', '); expected 127.0.0.1" }
    if (Get-FirewallRule) { throw 'the firewall rule is still there with LAN access off' }

    Write-Host '== uninstall removes the firewall rule'
    Invoke-Setup 'lan-again' @('/LANACCESS')
    Assert-LanAccess 'before uninstall'
    $uninstall = Start-Process -FilePath (Join-Path $InstallDir 'unins000.exe') -Wait -PassThru `
        -ArgumentList @('/VERYSILENT', '/SUPPRESSMSGBOXES', '/NORESTART', "/LOG=`"$uninstallLog`"", '/DELETEDATA')
    if ($uninstall.ExitCode -ne 0) { throw "the uninstaller exited with $($uninstall.ExitCode)" }
    Wait-Until { -not (Get-FirewallRule) } 60 'the firewall rule to be removed'

    Write-Host 'installer LAN access test OK'
} catch {
    Show-Diagnostics
    throw
} finally {
    if (Test-Path (Join-Path $InstallDir 'unins000.exe')) {
        Start-Process -FilePath (Join-Path $InstallDir 'unins000.exe') -Wait -ArgumentList '/VERYSILENT', '/SUPPRESSMSGBOXES', '/DELETEDATA'
    }
    Get-FirewallRule | Remove-NetFirewallRule
}
