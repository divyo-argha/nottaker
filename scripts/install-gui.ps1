$repo = "divyo-argha/octonote"
try {
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name.TrimStart('v')
} catch {
    $version = "1.0.1"
}
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") { "amd64" } else { "arm64" }

$cliBinary = "octonote-windows-$arch.exe"
$guiBinary = "octonote-gui-windows-$arch.exe"

$cliUrl = "https://github.com/$repo/releases/download/v$version/$cliBinary"
$guiUrl = "https://github.com/$repo/releases/download/v$version/$guiBinary"

$binDir = "$env:USERPROFILE\.octonote\bin"
if (!(Test-Path $binDir)) {
    New-Item -ItemType Directory -Force -Path $binDir | Out-Null
}

$cliDest = Join-Path $binDir "octonote.exe"
$guiDest = Join-Path $binDir "octonote-gui.exe"

Write-Host "Downloading octonote CLI & GUI..."
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
Write-Host "Run 'octonote' for the terminal interface."
Write-Host "Run 'octonote-gui' for the desktop application."
