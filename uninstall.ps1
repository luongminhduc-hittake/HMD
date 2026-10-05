# hittakeMD - Windows Uninstaller Script
$ErrorActionPreference = "SilentlyContinue"

Write-Host "=== Gỡ cài đặt hittakeMD ===" -ForegroundColor Yellow

$DesktopPath = [Environment]::GetFolderPath("Desktop")
Remove-Item "$DesktopPath\hittakeMD.lnk" -Force

$StartMenuPath = [Environment]::GetFolderPath("Programs")
Remove-Item "$StartMenuPath\hittakeMD.lnk" -Force

$InstallDir = "$env:LOCALAPPDATA\Programs\hittakeMD"
Remove-Item -Recurse -Force $InstallDir

Write-Host "✓ Đã gỡ bỏ hittakeMD khỏi máy tính." -ForegroundColor Green
