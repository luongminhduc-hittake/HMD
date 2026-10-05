# hittake's media downloader (hittakeMD)

> Công cụ tải video và âm thanh đa nền tảng mộc mạc, tinh gọn. Mình viết tool này cho bố mình để bố dễ dàng tải các bài nhạc, video, danh sách phát yêu thích về máy nghe offline.

---

## ✨ Điểm nổi bật

- **Tự động nhận link Clipboard**: Vừa mở app lên là link media đã được tự động điền sẵn, chỉ cần bấm `Enter` để tải.
- **Hỗ trợ đa nền tảng**: Tải mượt mà từ YouTube, TikTok, Facebook, Instagram, Twitter/X, SoundCloud, Bilibili...
- **1 file chạy duy nhất**: Tải về là dùng ngay (Windows `hmd.exe`, Linux `hmd`), nhúng sẵn Icon đẹp mắt, tự động chuẩn bị `yt-dlp` và `ffmpeg` trong nền.
- **Tích hợp sâu trên Windows**: Nhấp đúp chạy lần đầu sẽ tự động cài vào máy, tạo biểu tượng Desktop và Start Menu; tự động dọn sạch file `.old` sau khi cập nhật.
- **Tự động cập nhật**: Nhấn `u` trên giao diện khi có bản mới để tự nâng cấp trực tiếp.
- **7 Định dạng tối ưu**: MP3 (320kbps kèm bìa), M4A (gốc AAC), FLAC (Lossless), OPUS (Chất lượng cao), Video Best, 1080p, 720p.
- **Tùy chỉnh thư mục lưu**: Nhấn `c` trên giao diện để chọn thư mục lưu qua hộp thoại native hoặc nhập đường dẫn (mặc định: `Downloads/hittakeMD`).

---

## 📥 Tải về

Tải bản mới nhất tại [GitHub Releases](https://github.com/luongminhduc-hittake/HMD/releases/latest):
- **Windows**: `hmd-windows-amd64.exe` (hoặc `hmd.exe`)
- **Linux**: `hmd-linux-amd64` (hoặc `hmd`)

---

## 🚀 Cách dùng

1. Sao chép đường link video/nhạc trên YouTube, TikTok, Facebook...
2. Mở `hmd` (nhấp đúp trên Windows hoặc gõ `hmd` trên Linux) — link vừa copy sẽ được tự động điền sẵn!
3. Nhấn `Enter`. Có thể bấm `c` nếu muốn đổi thư mục lưu.
4. Bấm số từ `1` đến `7` để chọn định dạng tải.
5. Nhạc và video được lưu tự động tại `Downloads/hittakeMD`.

---

## 🪟 Cài đặt trên Windows

- **Cách 1**: Chỉ cần nhấp đúp file `hmd.exe`, ứng dụng sẽ tự động tạo Shortcut ra Màn hình chính (Desktop) và Start Menu.
- **Cách 2**: Chạy script PowerShell:
  ```powershell
  powershell -ExecutionPolicy Bypass -File .\install.ps1
  ```

---

## 🐧 Cài đặt trên Linux

```bash
./install.sh    # Tự động cài vào ~/.local/bin/hmd, tạo symlink ytdl và thêm icon vào App Menu
./uninstall.sh  # Gỡ cài đặt
```

---

## 🛠 Biên dịch từ mã nguồn

```bash
go build -ldflags="-s -w" -o hmd .                                # Linux
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o hmd.exe .   # Windows
```
