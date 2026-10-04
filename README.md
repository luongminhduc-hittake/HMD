# hittake's ytdl (ytdownloader)

> Công cụ tải video và âm thanh từ YouTube mộc mạc, tinh gọn. Mình viết tool này cho bố mình để bố dễ dàng tải các bài nhạc, danh sách bài hát yêu thích về máy nghe offline.

---

## ✨ Điểm nổi bật

- **1 file chạy duy nhất**: Tải về là dùng ngay (Windows `ytdl.exe`, Linux `ytdl`), tự động chuẩn bị `yt-dlp` và `ffmpeg` trong nền.
- **Tự động cập nhật**: Nhấn `u` trên giao diện khi có bản mới để tự nâng cấp trực tiếp.
- **Tải Playlist thông minh**: Tự động bỏ qua video bị ẩn hoặc không khả dụng, không làm gián đoạn tiến trình tải.
- **5 Định dạng tối ưu**: MP3 (320kbps kèm ảnh bìa), M4A (gốc AAC), Video Best, 1080p, 720p.

---

## 📥 Tải về

Tải bản mới nhất tại [GitHub Releases](https://github.com/luongminhduc-hittake/ytdownloader/releases/latest):
- **Windows**: `ytdl-windows-amd64.exe` (hoặc `ytdl.exe`)
- **Linux**: `ytdl-linux-amd64` (hoặc `ytdl`)

---

## 🚀 Cách dùng

1. Sao chép đường link video hoặc playlist trên YouTube.
2. Mở `ytdl` (nhấp đúp trên Windows hoặc gõ `ytdl` trên Linux).
3. Dán link (`Ctrl + V`), nhấn `Enter`.
4. Bấm số từ `1` đến `5` để chọn định dạng tải.
5. Nhạc và video được lưu tự động tại thư mục `Downloads/YouTube`.

---

## 🐧 Cài đặt trên Linux

```bash
./install.sh    # Tự động cài vào ~/.local/bin/ytdl và thêm icon vào App Menu
./uninstall.sh  # Gỡ cài đặt
```

---

## 🛠 Biên dịch từ mã nguồn

```bash
go build -ldflags="-s -w" -o ytdl .                                # Linux
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ytdl.exe .   # Windows
```
