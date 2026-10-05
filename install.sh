#!/usr/bin/env bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BIN_DIR="${HOME}/.local/bin"
APP_DIR="${HOME}/.local/share/applications"
HICOLOR_DIR="${HOME}/.local/share/icons/hicolor"
PIXMAPS_DIR="${HOME}/.local/share/pixmaps"

echo "⚡ Đang biên dịch hittakeMD (hmd) cho Linux..."
cd "${SCRIPT_DIR}"
go build -ldflags="-s -w" -o hmd .

echo "📦 Đang cài đặt vào ${BIN_DIR}..."
mkdir -p "${BIN_DIR}" "${APP_DIR}" "${HICOLOR_DIR}/256x256/apps" "${HICOLOR_DIR}/scalable/apps" "${PIXMAPS_DIR}"
install -m 755 "${SCRIPT_DIR}/hmd" "${BIN_DIR}/hmd"
rm -f "${BIN_DIR}/ytdl" "${APP_DIR}/ytdl.desktop"

echo "🎨 Đang cài đặt icon ứng dụng vào icon theme (hicolor & pixmaps)..."
install -m 644 "${SCRIPT_DIR}/assets/icon.png" "${HICOLOR_DIR}/256x256/apps/hittakeMD.png"
install -m 644 "${SCRIPT_DIR}/assets/icon.png" "${HICOLOR_DIR}/256x256/apps/hmd.png"
install -m 644 "${SCRIPT_DIR}/assets/icon.svg" "${HICOLOR_DIR}/scalable/apps/hittakeMD.svg"
install -m 644 "${SCRIPT_DIR}/assets/icon.svg" "${HICOLOR_DIR}/scalable/apps/hmd.svg"
install -m 644 "${SCRIPT_DIR}/assets/icon.png" "${PIXMAPS_DIR}/hittakeMD.png"
install -m 644 "${SCRIPT_DIR}/assets/icon.png" "${PIXMAPS_DIR}/hmd.png"

echo "🖥  Đang cài đặt desktop entry vào ${APP_DIR}..."
install -m 644 "${SCRIPT_DIR}/hmd.desktop" "${APP_DIR}/hmd.desktop"

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${APP_DIR}" >/dev/null 2>&1 || true
fi

if command -v omarchy-menu >/dev/null 2>&1; then
    omarchy-menu refresh >/dev/null 2>&1 || true
fi

echo "✔ Cài đặt hittakeMD hoàn tất thành công!"
echo "  - Lệnh gõ:  ${BIN_DIR}/hmd"
echo "  - Desktop:  ${APP_DIR}/hmd.desktop"
echo "  - Icon:     ${HICOLOR_DIR}/256x256/apps/hittakeMD.png"

if [[ ":$PATH:" != *":${BIN_DIR}:"* ]]; then
    echo "⚠ Lưu ý: ${BIN_DIR} chưa nằm trong biến môi trường PATH của bạn."
    echo "  Hãy thêm dòng sau vào ~/.bashrc hoặc ~/.zshrc:"
    echo "  export PATH=\"\$HOME/.local/bin:\$PATH\""
fi
