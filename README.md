# ytdl - Modern YouTube Downloader for Windows

Ứng dụng CLI / TUI độc lập cho Windows tải video và âm thanh từ YouTube, thiết kế theo phong cách Cyber/Nord tinh tế, mượt mà và tối giản. Chỉ một file thực thi duy nhất `ytdl.exe`.

---

## ✨ Điểm nổi bật

- **Chỉ 1 file thực thi duy nhất (`ytdl.exe`)**: Dung lượng siêu nhẹ (~7.4 MB), không cần cài đặt Python hay Node.js.
- **Tự động quản lý công cụ (`yt-dlp` & `ffmpeg`)**: Nếu hệ thống chưa có, ứng dụng sẽ tự động tải các bản portable chính thức vào `%LOCALAPPDATA%\ytdl\bin` với thanh tiến trình trực quan trong lần chạy đầu tiên.
- **Chế độ kép (Hybrid Mode)**:
  - **Double-click hoặc chạy không cờ**: Mở giao diện tương tác TUI (Bubble Tea & Lipgloss) hiện đại.
  - **Dòng lệnh (CLI Flags)**: Hỗ trợ chạy script tự động nhanh gọn.
- **Hỗ trợ 5 Preset tối ưu**:
  - 🎥 **Best Video (MP4)**: Tự động gộp video độ phân giải cao nhất (4K/2K/1080p) + audio tốt nhất.
  - 📺 **Full HD 1080p (MP4)**
  - 💻 **HD 720p (MP4)**
  - 🎵 **Audio MP3 (320kbps)**: Tự động trích xuất, nhúng ảnh bìa (thumbnail) và ID3 metadata.
  - ⚡ **Audio M4A gốc (AAC)**: Tải trực tiếp không cần chuyển mã, siêu tốc.
- **Tự động phát hiện Playlist**: Cho phép chọn tải video đơn lẻ hay toàn bộ danh sách phát.
- **Menu hành động sau khi tải**: Hiển thị thông số file, hỗ trợ bấm mở nhanh thư mục lưu bằng Windows Explorer.

---

## 🚀 Hướng dẫn sử dụng

### 1. Chế độ Tương tác (TUI)
Nhấp đúp chuột vào file `ytdl.exe` hoặc mở Command Prompt / PowerShell:
```cmd
ytdl.exe
```
1. Dán đường dẫn video hoặc playlist (`Ctrl+V` hoặc chuột phải).
2. Dùng phím mũi tên `↑` / `↓` hoặc phím số `1-5` để chọn chất lượng tải.
3. Nhấn `Enter` để bắt đầu tải với thanh tiến trình thời gian thực (tốc độ, ETA, dung lượng).
4. Sau khi tải xong, chọn **[Mở thư mục chứa file]** hoặc **[Tải tiếp video khác]**.

---

### 2. Chế độ Dòng lệnh (CLI)
Dành cho người dùng chuyên nghiệp hoặc chạy tự động:
```cmd
# Tải video chất lượng tốt nhất
ytdl.exe "https://www.youtube.com/watch?v=..."

# Tải nhạc MP3 (320kbps + bìa)
ytdl.exe -f mp3 "https://www.youtube.com/watch?v=..."

# Tải video 1080p và lưu vào thư mục mong muốn
ytdl.exe -f 1080p -o "D:\Videos" "https://www.youtube.com/watch?v=..."

# Tải toàn bộ playlist
ytdl.exe --playlist "https://www.youtube.com/playlist?list=..."

# Cập nhật yt-dlp lên bản mới nhất
ytdl.exe --update
```

---

## 🛠 Biên dịch từ mã nguồn

Yêu cầu: Go 1.22+

```bash
# Biên dịch cho Windows x64 (tạo file ytdl.exe)
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ytdl.exe .

# Chạy kiểm thử tự động
go test -v ./...
```
