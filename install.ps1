$ErrorActionPreference = "Stop"

Write-Host "Installing claude-switch..."

$ScriptDir = $PSScriptRoot
if (-not $ScriptDir) {
    $ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
}
Set-Location $ScriptDir

go build -o claude-switch.exe .\cmd\claude-switch

$InstallDir = "$env:USERPROFILE\.local\bin"
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}
Move-Item -Force claude-switch.exe "$InstallDir\claude-switch.exe"

Write-Host "Installed to $InstallDir\claude-switch.exe"
Write-Host "Add $InstallDir to your PATH if not already added"
