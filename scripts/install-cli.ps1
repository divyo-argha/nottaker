$repo = "divyo-argha/nottaker"
$version = "1.0.0"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }
$binary = "nottaker-windows-$arch.exe"
$url = "https://github.com/$repo/releases/download/v$version/$binary"

$binDir = "$env:USERPROFILE\.nottaker\bin"
if (!(Test-Path $binDir)) {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
}

$destPath = Join-Path $binDir "nottaker.exe"
Write-Host "Downloading nottaker CLI from $url..."
Invoke-WebRequest -Uri $url -OutFile $destPath

# Add path to User Env Path if not present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", "User")
    $env:Path += ";$binDir"
    Write-Host "Added $binDir to your user PATH variable."
}

Write-Host "nottaker CLI successfully installed to $destPath"
Write-Host "Please restart your terminal to use the command 'nottaker'."
