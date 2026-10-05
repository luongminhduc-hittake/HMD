#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/.local/bin"
APP_DIR="${HOME}/.local/share/applications"
ICON_DIR="${HOME}/.local/share/icons"

echo "⚡ Đang biên dịch hittakeMD (hmd) cho Linux..."
cd "${SCRIPT_DIR}"
go build -ldflags="-s -w" -o hmd .

echo "📦 Đang cài đặt vào ${BIN_DIR}..."
mkdir -p "${BIN_DIR}" "${APP_DIR}" "${ICON_DIR}"
install -m 755 "${SCRIPT_DIR}/hmd" "${BIN_DIR}/hmd"
ln -sf "${BIN_DIR}/hmd" "${BIN_DIR}/ytdl"

echo "🎨 Đang cài đặt icon ứng dụng vào ${ICON_DIR}..."
install -m 644 "${SCRIPT_DIR}/assets/icon.png" "${ICON_DIR}/hittakeMD.png"

echo "🖥  Đang cài đặt desktop entry vào ${APP_DIR}..."
install -m 644 "${SCRIPT_DIR}/hmd.desktop" "${APP_DIR}/hmd.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${APP_DIR}" >/dev/null 2>&1 || true
fi

echo "✔ Cài đặt hittakeMD hoàn tất thành công!"
echo "  - Lệnh gõ:  ${BIN_DIR}/hmd (hoặc gõ ytdl)"
echo "  - Desktop:  ${APP_DIR}/hmd.desktop"
echo "  - Icon:     ${ICON_DIR}/hittakeMD.png"

if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    echo "⚠ Lưu ý: ${BIN_DIR} chưa nằm trong biến môi trường PATH của bạn."
    echo "  Hãy thêm dòng sau vào ~/.bashrc hoặc ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
