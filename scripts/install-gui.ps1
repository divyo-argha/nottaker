$repo = "nottaker/nottaker"
$version = "0.1.0"
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }

$cliBinary = "nottaker-windows-$arch.exe"
$guiBinary = "nottaker-gui-windows-$arch.exe"

$cliUrl = "https://github.com/$repo/releases/download/v$version/$cliBinary"
$guiUrl = "https://github.com/$repo/releases/download/v$version/$guiBinary"

$binDir = "$env:USERPROFILE\.nottaker\bin"
if (!(Test-Path $binDir)) {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
}

$cliDest = Join-Path $binDir "nottaker.exe"
$guiDest = Join-Path $binDir "nottaker-gui.exe"

Write-Host "Downloading nottaker CLI & GUI..."
Invoke-WebRequest -Uri $cliUrl -OutFile $cliDest
Invoke-WebRequest -Uri $guiUrl -OutFile $guiDest

# Add path to User Env Path if not present
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binDir", "User")
    $env:Path += ";$binDir"
    Write-Host "Added $binDir to your user PATH variable."
}

Write-Host "✓ Installation complete!"
Write-Host "Please restart your terminal to use the new commands."
Write-Host "Run 'nottaker' for the terminal interface."
Write-Host "Run 'nottaker-gui' for the desktop application."
