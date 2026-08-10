# 🤖 Backend - Projek Bot WhatsApp

Ini adalah bagian inti server (Backend) untuk Bot WhatsApp. Aplikasi ini bertugas mengelola koneksi WhatsApp (menggunakan pustaka `whatsmeow`), menyediakan REST API untuk antarmuka pengguna (Frontend), serta mengatur integrasi lain seperti sinkronisasi dengan Google Sheets.

## 🚀 Getting Started

### 📋 Prasyarat

Pastikan Anda sudah menginstal:
- [Go (Golang)](https://go.dev/dl/) versi 1.19 atau lebih baru.
- (Opsional) Docker & Docker Compose jika ingin menjalankan menggunakan kontainer.

### 📦 Konfigurasi `.env`

1. Salin file `.env.example` menjadi `.env`.
2. Sesuaikan kredensial di dalamnya:
   ```bash
   cp .env.example .env
   ```

### 🧪 Run the App (Lokal)

Jalankan perintah berikut untuk mengunduh *dependencies* dan menjalankan server:

```bash
go mod tidy
go run cmd/app/main.go
```
Server akan berjalan di port `8080`.

### 🐳 Run dengan Docker Compose

Jika Anda ingin menjalankan Backend, Frontend, dan Redis secara bersamaan:

```bash
docker-compose up -d --build
```

## 🌐 Tech Stack

- 🐹 **Golang** - Bahasa pemrograman utama.
- 🚀 **Fiber (v2)** - Framework web yang sangat cepat untuk Golang.
- 💬 **Whatsmeow** - Library utama untuk menghubungkan bot ke server WhatsApp.
- 🗄️ **GORM & PostgreSQL** - ORM dan Database yang digunakan.
- 📝 **Google Sheets API** - Untuk integrasi data ke Spreadsheet.
- 🐋 **Docker** - Manajemen kontainer.

## 📂 Struktur Folder Utama

- `cmd/app/` - Titik masuk (*entry point*) aplikasi (`main.go`).
- `internal/` - Berisi logika inti aplikasi yang tidak diekspos keluar (handler, model, config, provider).
- `data/` - Folder tempat menyimpan sesi lokal WhatsApp (`wa_session.db`).
