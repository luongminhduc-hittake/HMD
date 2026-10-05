# hittakeMD - Windows Installer Script
$ErrorActionPreference = "Stop"

Write-Host "=== Cài đặt hittakeMD (hittake's media downloader) ===" -ForegroundColor Cyan

$InstallDir = "$env:LOCALAPPDATA\Programs\hittakeMD"
$TargetExe = "$InstallDir\hmd.exe"

New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$CurrentDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$SourceExe = "$CurrentDir\hmd.exe"

if (Test-Path $SourceExe) {
    Copy-Item -Path $SourceExe -Destination $TargetExe -Force
} else {
    Write-Host "Đang tải bản hmd.exe mới nhất từ GitHub..." -ForegroundColor Yellow
    $DownloadUrl = "https://github.com/luongminhduc-hittake/HMD/releases/latest/download/hmd-windows-amd64.exe"
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetExe -UseBasicParsing
}

# Tạo phím tắt Desktop & Start Menu
$WshShell = New-Object -ComObject WScript.Shell
$DesktopPath = [Environment]::GetFolderPath("Desktop")
$ShortcutDesktop = $WshShell.CreateShortcut("$DesktopPath\hittakeMD.lnk")
$ShortcutDesktop.TargetPath = $TargetExe
$ShortcutDesktop.WorkingDirectory = $InstallDir
$ShortcutDesktop.Description = "hittake's media downloader"
$ShortcutDesktop.Save()

$StartMenuPath = [Environment]::GetFolderPath("Programs")
$ShortcutStart = $WshShell.CreateShortcut("$StartMenuPath\hittakeMD.lnk")
$ShortcutStart.TargetPath = $TargetExe
$ShortcutStart.WorkingDirectory = $InstallDir
$ShortcutStart.Description = "hittake's media downloader"
$ShortcutStart.Save()

Write-Host "✓ Đã cài đặt hittakeMD thành công vào: $InstallDir" -ForegroundColor Green
Write-Host "✓ Đã tạo biểu tượng Desktop và Start Menu!" -ForegroundColor Green
