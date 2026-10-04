# ytdl - Modern YouTube Downloader (Windows & Linux)

Ứng dụng CLI / TUI độc lập tải video và âm thanh từ YouTube, thiết kế theo phong cách Cyber/Nord tinh tế, mượt mà và tối giản. Hỗ trợ cả **Windows** (1 file exe) và **Linux** (tích hợp menu `.desktop`).

---

## ✨ Điểm nổi bật

- **Chỉ 1 file thực thi duy nhất**:
  - Windows: `ytdl.exe` (~7.4 MB).
  - Linux: `ytdl` (~7.1 MB).
- **Tự động Cập nhật (Self-Update)**: Tự động kiểm tra phiên bản mới từ GitHub Releases khi khởi động; người dùng chỉ cần nhấn `u` trên giao diện để tự động nâng cấp file thực thi.
- **Tải Playlist thông minh**: Tự động bỏ qua các video bị ẩn, private hoặc không khả dụng trong danh sách phát mà không làm gián đoạn tiến trình tải.
- **Tự động quản lý công cụ (`yt-dlp` & `ffmpeg`)**: Nếu hệ thống chưa có, ứng dụng sẽ tự động tải các bản portable chính thức vào `%LOCALAPPDATA%\ytdl\bin` (Windows) hoặc `~/.local/share/ytdl/bin` (Linux) với thanh tiến trình trực quan trong lần chạy đầu tiên.
- **Tích hợp Linux Desktop Entry (`.desktop`)**: Tự động hiển thị trên menu ứng dụng (Rofi, Wofi, GNOME, KDE, Omarchy, ...) với icon hệ thống và cờ `Terminal=true`.
- **Giao diện TUI hiện đại**: Xây dựng trên Bubble Tea & Lipgloss, trực quan, hỗ trợ phím tắt và chuột.
- **Hỗ trợ 5 Preset tối ưu**:
  - 🎥 **Best Video (MP4)**: Tự động gộp video độ phân giải cao nhất (4K/2K/1080p) + audio tốt nhất.
  - 📺 **Full HD 1080p (MP4)**
  - 💻 **HD 720p (MP4)**
  - 🎵 **Audio MP3 (320kbps)**: Tự động trích xuất, nhúng ảnh bìa (thumbnail) và ID3 metadata.
  - ⚡ **Audio M4A gốc (AAC)**: Tải trực tiếp không cần chuyển mã, siêu tốc.
- **Menu hành động sau khi tải**: Hiển thị thông số file, hỗ trợ bấm mở nhanh thư mục lưu bằng File Manager.

---

## 🐧 Hướng dẫn Cài đặt & Sử dụng trên Linux

### 1. Cài đặt nhanh vào hệ thống (không cần sudo)
```bash
./install.sh
```
Lệnh trên sẽ:
- Biên dịch binary `ytdl` vào `~/.local/bin/ytdl`.
- Cài đặt `ytdl.desktop` vào `~/.local/share/applications/ytdl.desktop`.
- Cập nhật cơ sở dữ liệu desktop để ứng dụng xuất hiện ngay trong Menu / App Launcher.

### 2. Gỡ cài đặt
```bash
./uninstall.sh
```

### 3. Khởi chạy
- **Từ Menu ứng dụng**: Tìm `YouTube Downloader` hoặc `ytdl` trong App Launcher / Rofi.
- **Từ Terminal (TUI)**: Gõ `ytdl` (hoặc `ytdl "<link>"` để dán sẵn link).

---

## 🪟 Hướng dẫn Sử dụng trên Windows

1. Tải hoặc copy file **`ytdl.exe`** vào máy.
2. Nhấp đúp chuột để mở giao diện tương tác (TUI), dán link và chọn định dạng tải.
3. Khi có bản cập nhật mới, giao diện sẽ hiện thông báo; chỉ cần nhấn phím `u` để app tự nâng cấp.

---

## 🛠 Biên dịch từ mã nguồn

Yêu cầu: Go 1.22+

```bash
# Biên dịch cho Linux (tạo file ytdl)
go build -ldflags="-s -w" -o ytdl .

# Biên dịch chéo cho Windows x64 (tạo file ytdl.exe)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ytdl.exe .

# Chạy kiểm thử tự động
go test -v ./...
```
