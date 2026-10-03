#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/.local/bin"
APP_DIR="${HOME}/.local/share/applications"

echo "⚡ Đang biên dịch ytdl cho Linux..."
cd "${SCRIPT_DIR}"
go build -ldflags="-s -w" -o ytdl .

echo "📦 Đang cài đặt vào ${BIN_DIR}..."
mkdir -p "${BIN_DIR}" "${APP_DIR}"
install -m 755 "${SCRIPT_DIR}/ytdl" "${BIN_DIR}/ytdl"

echo "🖥  Đang cài đặt desktop entry vào ${APP_DIR}..."
install -m 644 "${SCRIPT_DIR}/ytdl.desktop" "${APP_DIR}/ytdl.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${APP_DIR}" >/dev/null 2>&1 || true
fi

echo "✔ Cài đặt hoàn tất thành công!"
echo "  - Binary:  ${BIN_DIR}/ytdl"
echo "  - Desktop: ${APP_DIR}/ytdl.desktop"

if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    echo "⚠ Lưu ý: ${BIN_DIR} chưa nằm trong biến môi trường PATH của bạn."
    echo "  Hãy thêm dòng sau vào ~/.bashrc hoặc ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
