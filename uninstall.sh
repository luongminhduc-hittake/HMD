#!/usr/bin/env bash
set -e

BIN_FILE="${HOME}/.local/bin/ytdl"
DESKTOP_FILE="${HOME}/.local/share/applications/ytdl.desktop"

echo "🗑  Đang gỡ cài đặt ytdl..."

if [ -f "${BIN_FILE}" ]; then
    rm -f "${BIN_FILE}"
    echo "✔ Đã xóa ${BIN_FILE}"
fi

if [ -f "${DESKTOP_FILE}" ]; then
    rm -f "${DESKTOP_FILE}"
    echo "✔ Đã xóa ${DESKTOP_FILE}"
fi

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${HOME}/.local/share/applications" >/dev/null 2>&1 || true
fi

echo "✔ Gỡ cài đặt hoàn tất!"
