# hittake's ytdl (ytdownloader)

> **Công cụ tải video và âm thanh từ YouTube mộc mạc, tinh gọn. Mình viết tool này cho bố mình để bố dễ dàng tải các bài nhạc, danh sách bài hát yêu thích về máy nghe offline mà không phải mò mẫm qua mấy trang web tải nhạc đầy quảng cáo rác phức tạp.**

---

## ✨ Điểm nổi bật

- **Cực kỳ đơn giản, chỉ 1 file chạy duy nhất**:
  - Windows: `ytdl.exe` (~7.4 MB).
  - Linux: `ytdl` (~7.8 MB).
- **Tự động Cập nhật (Self-Update)**: Mỗi khi mở app, chương trình sẽ tự kiểm tra xem có bản mới hay không. Chỉ cần bấm phím `u` là app tự động nâng cấp, không cần phải tải lại file thủ công.
- **Tải Playlist thông minh**: Tự động bỏ qua các video bị ẩn, riêng tư hoặc không khả dụng trong danh sách phát; tiếp tục tải trọn vẹn toàn bộ bài hát còn lại mà không bao giờ bị đứng máy hay báo lỗi nửa chừng.
- **Tự lo mọi công cụ phụ trợ (`yt-dlp` & `ffmpeg`)**: Lần đầu chạy nếu máy chưa có sẵn, app sẽ tự tải các bản portable chính thức vào máy trong nền kèm thanh phần trăm trực quan. Người dùng không cần cài đặt thêm bất kỳ phần mềm nào khác.
- **Giao diện dòng lệnh trực quan (TUI)**: Thiết kế tông màu Cyber/Nord dịu mắt, dễ nhìn, thao tác hoàn toàn bằng phím số (1, 2, 3, 4, 5) hoặc phím mũi tên.
- **5 Định dạng chọn sẵn tối ưu**:
  - 🎵 **Audio MP3 (320kbps)**: Trích xuất nhạc chất lượng cao, tự động gắn ảnh bìa bài hát và thông tin bài.
  - ⚡ **Audio M4A gốc (AAC)**: Tải stream âm thanh trực tiếp siêu tốc, không cần convert.
  - 🎥 **Best Video (MP4)**: Tự ghép hình ảnh nét nhất (4K/2K/1080p) và âm thanh hay nhất.
  - 📺 **Full HD 1080p (MP4)**: Tương thích chuẩn mọi tivi, máy tính, điện thoại.
  - 💻 **HD 720p (MP4)**: Tiết kiệm dung lượng bộ nhớ.
- **Mở nhanh sau khi tải**: Tải xong có sẵn tùy chọn bấm phím để mở ngay thư mục chứa file nhạc vừa tải về.

---

## 📥 Tải về (Releases)

Truy cập mục [Releases](https://github.com/luongminhduc-hittake/ytdownloader/releases/latest) để tải bản biên dịch sẵn mới nhất:
- **Windows**: Tải file `ytdl-windows-amd64.exe` (hoặc `ytdl.exe`).
- **Linux**: Tải file `ytdl-linux-amd64` (hoặc `ytdl`).

---

## 🚀 Hướng dẫn Sử dụng Nhanh

1. **Sao chép link**: Mở YouTube trên trình duyệt hoặc điện thoại, sao chép đường link bài hát/video (hoặc cả danh sách phát Playlist).
2. **Mở ytdl**:
   - Trên **Windows**: Nhấp đúp chuột vào file `ytdl.exe`.
   - Trên **Linux**: Mở từ menu ứng dụng hoặc gõ lệnh `ytdl` trong Terminal.
3. **Dán link**: Bấm tổ hợp phím `Ctrl + V` (hoặc nhấp chuột phải) để dán link rồi bấm `Enter`.
4. **Chọn định dạng**: Bấm các phím số từ `1` đến `5` (hoặc dùng mũi tên `↑` `↓` rồi nhấn `Enter`).
5. **Thưởng thức**: Chờ app tải xong, chọn `Mở thư mục chứa file` để nghe bài hát vừa tải.
   - Thư mục lưu mặc định: `Downloads/YouTube` trong máy.

---

## 🐧 Cài đặt & Sử dụng trên Linux

### 1. Cài đặt vào hệ thống (không cần quyền root / sudo)
```bash
./install.sh
```
Script sẽ tự động:
- Biên dịch `ytdl` vào thư mục `~/.local/bin/ytdl`.
- Tạo biểu tượng desktop entry `~/.local/share/applications/ytdl.desktop` để tìm thấy app ngay trong Menu ứng dụng (Rofi, Wofi, GNOME, KDE, Hyprland/Omarchy...).

### 2. Gỡ cài đặt
```bash
./uninstall.sh
```

### 3. Khởi chạy
- **Từ Menu ứng dụng**: Tìm `YouTube Downloader` hoặc `ytdl`.
- **Từ Terminal**: Gõ `ytdl` (có thể truyền sẵn link: `ytdl "<link>"`).

---

## 🪟 Sử dụng trên Windows

1. Tải file `ytdl.exe` về máy (khuyên dùng: để ngoài Desktop hoặc thư mục riêng dễ nhớ).
2. Nhấp đúp chuột để mở giao diện, dán link và chọn định dạng tải.
3. Khi có bản cập nhật mới, giao diện sẽ xuất hiện thông báo màu vàng; chỉ cần nhấn phím `u` là app tự nâng cấp.

---

## 🛠 Biên dịch từ mã nguồn

Yêu cầu môi trường: **Go 1.22+**

```bash
# Clone dự án
git clone https://github.com/luongminhduc-hittake/ytdownloader.git
cd ytdownloader

# Biên dịch cho Linux
go build -ldflags="-s -w" -o ytdl .

# Biên dịch chéo cho Windows x64
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ytdl.exe .

# Chạy kiểm thử tự động
go test -v ./...
```
