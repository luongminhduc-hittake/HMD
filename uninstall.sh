#!/usr/bin/env bash
set -e

echo "🗑  Đang gỡ cài đặt hittakeMD..."

for bin in "${HOME}/.local/bin/hmd" "${HOME}/.local/bin/ytdl"; do
    if [ -f "${bin}" ] || [ -L "${bin}" ]; then
        rm -f "${bin}"
        echo "✔ Đã xóa ${bin}"
    fi
done

for desktop in "${HOME}/.local/share/applications/hmd.desktop" "${HOME}/.local/share/applications/ytdl.desktop"; do
    if [ -f "${desktop}" ]; then
        rm -f "${desktop}"
        echo "✔ Đã xóa ${desktop}"
    fi
done

if [ -f "${HOME}/.local/share/icons/hittakeMD.png" ]; then
    rm -f "${HOME}/.local/share/icons/hittakeMD.png"
    echo "✔ Đã xóa icon ${HOME}/.local/share/icons/hittakeMD.png"
fi

if command -v update-desktop-database >/dev/null 2>&1; then
    update-desktop-database "${HOME}/.local/share/applications" >/dev/null 2>&1 || true
fi

echo "✔ Gỡ cài đặt hittakeMD hoàn tất!"
