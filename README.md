# 🤖 Backend Microservice - Go Fiber WhatsApp Bot & E-Commerce API

Microservice API Gateway & Core Engine berbasis **Golang (Fiber v2)** & **GORM** untuk mengelola layanan E-Commerce, Autentikasi User/Admin, WhatsApp Webhook Engine (WAHA), serta komunikasi real-time via WebSockets.

---

## 📋 Daftar Isi
- [🏗️ Arsitektur & Peran Microservice](#️-arsitektur--peran-microservice)
- [🌐 Tech Stack Backend](#-tech-stack-backend)
- [📁 Struktur Folder Backend (`internal/`)](#-struktur-folder-backend-internal)
- [🗄️ Model Data GORM (`internal/model/`)](#️-model-data-gorm-internalmodel)
- [📡 API Documentation Endpoints](#-api-documentation-endpoints)
- [⚙️ Pengaturan Environment Variables (`.env`)](#️-pengaturan-environment-variables-env)
- [🧪 Cara Menjalankan Engine](#-cara-menjalankan-engine)

---

## 🏗️ Arsitektur & Peran Microservice

Microservice ini menangani seluruh proses bisnis kritis pada sistem:
1. 🛍️ **Core E-Commerce REST API**: Layanan CRUD Produk, Kategori, Banner Promo, dan Dashboard Stats.
2. 💬 **WhatsApp Gateway & Webhook Engine**:
   - Integrasi WAHA (WhatsApp HTTP API Provider) untuk otentikasi QR Code dan kontrol sesi.
   - Penampung Webhook pesan masuk real-time (`/api/wa/webhook`).
   - Fitur Live Chat & Manajemen Riwayat Pesan (`/api/chat/...`).
3. ⚡ **Real-Time WebSocket Hub (`wshub`)**: Penyiaran (broadcasting) pesan baru dan event WhatsApp secara instan ke frontend.
4. 🔐 **Role-Based Authentication**: Sistem autentikasi terpisah untuk Admin dan Customer (JWT & Password Hashing).
5. 📊 **Auto Database Seeding**: Otomatis mengisi kategori awal (`Computer & Laptop`, `Gaming Console`, `Smartphone`, `Headphone`, `Computer Accessories`, `Umum`) jika database kosong.

---

## 🌐 Tech Stack Backend

- 🐹 **Golang 1.20+** - Language Engine dengan konkurensi goroutine.
- ⚡ **Fiber v2 (`github.com/gofiber/fiber/v2`)** - Express-inspired HTTP framework tercepat untuk Go.
- 🗄️ **GORM (`gorm.io/gorm`)** - ORM Database (SQLite / PostgreSQL / MySQL driver).
- 🔌 **Gorilla WebSockets (`github.com/gofiber/websocket/v2`)** - Komunikasi dua arah real-time.
- 💬 **WAHA (WhatsApp HTTP API)** - Provider sesi WhatsApp Web.
- 🐳 **Docker & Docker Compose** - Kontainerisasi produksi.

---

## 📁 Struktur Folder Backend (`internal/`)

```text
Bot_BE/
├── cmd/
│   └── main.go                      # Entrypoint utama server
├── internal/
│   ├── config/
│   │   └── config.go                # Loader .env & variabel sistem
│   ├── provider/
│   │   ├── database.go              # Inisialisasi GORM, Auto-Migration, & DB Driver
│   │   ├── sheets.go                # Integrasi Google Sheets API
│   │   └── waha.go                  # Client integrasi WAHA WhatsApp Engine
│   ├── model/                       # Skema Struct Database GORM
│   │   ├── product.go               # Struct Model Produk
│   │   ├── category.go              # Struct Model Kategori Produk
│   │   ├── banner.go                # Struct Model Banner Promo
│   │   ├── customer.go              # Struct Model Pelanggan (Customer)
│   │   ├── admin.go                 # Struct Model Administrator
│   │   ├── config.go                # Struct Model Konfigurasi Bot
│   │   ├── activity_log.go          # Struct Model Audit Log Aktivitas
│   │   └── models.go                # Response DTO Global
│   ├── handler/                     # HTTP Handlers (REST Controllers)
│   │   ├── product_handler.go       # Controller CRUD & Search Produk
│   │   ├── category_handler.go      # Controller CRUD Kategori & Auto-Seed
│   │   ├── banner_handler.go        # Controller CRUD Banner Promo
│   │   ├── auth_handler.go          # Controller Login/Auth Admin
│   │   ├── customer_handler.go      # Controller Login/Register Customer
│   │   ├── message_handler.go       # Controller Webhook WA & Broadcast
│   │   ├── chat_handler.go          # Controller Live Chat History & Send
│   │   ├── config_handler.go        # Controller Spreadsheet Config
│   │   └── activity_handler.go      # Controller Log Aktivitas
│   ├── server/
│   │   └── server.go                # Setup Fiber Routes, CORS, & WS Upgrade
│   ├── service/
│   │   └── bot_service.go           # Logika Bisnis Bot & Sheet Processing
│   └── wshub/
│       └── hub.go                   # Connection Hub WebSockets
├── data/                            # Database File Storage (bot.db)
├── Dockerfile                       # Production Multi-Stage Dockerfile
├── docker-compose.yml               # Docker Compose Orchestration
└── go.mod                           # Go Module Manifest
```

---

## 🗄️ Model Data GORM (`internal/model/`)

### 1. Model Produk (`product.go`)
```go
type Product struct {
	ID                uint    `gorm:"primaryKey" json:"id"`
	Title             string  `json:"title"`
	SKU               string  `json:"sku"`
	Brand             string  `json:"brand"`
	Category          string  `json:"category"`
	ShortDescription  string  `json:"short_description"`
	Description       string  `json:"description"`
	Price             float64 `json:"price"`
	DiscountPrice     float64 `json:"discount_price"`
	Stock             int     `json:"stock"`
	LowStockThreshold int     `json:"low_stock_threshold"`
	Weight            float64 `json:"weight"`
	Image             string  `json:"image"`
	Status            string  `json:"status"`
	IsFeatured        bool    `json:"is_featured"`
	IsActive          bool    `json:"is_active"`
}
```

### 2. Model Kategori (`category.go`)
```go
type Category struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	IsActive    bool   `json:"is_active"`
}
```

---

## 📡 API Documentation Endpoints

### 🛍️ 1. Produk API (`/api/products`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/products` | mengambil seluruh daftar produk |
| `GET` | `/api/products/search?q=macbook` | pencarian cepat produk berdasarkan keyword/SKU |
| `GET` | `/api/products/:id` | mengambil detail produk berdasarkan ID |
| `POST` | `/api/products` | membuat produk baru |
| `POST` | `/api/products/bulk` | import produk massal via CSV/JSON |
| `PUT` | `/api/products/:id` | memperbarui data produk |
| `DELETE` | `/api/products/:id` | menghapus produk |

---

### 🏷️ 2. Kategori API (`/api/categories`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/categories` | mengambil daftar kategori (Otomatis auto-seed jika DB kosong) |
| `POST` | `/api/categories` | membuat kategori baru (Validasi nama duplikat HTTP 400) |
| `PUT` | `/api/categories/:id` | mengedit data kategori |
| `DELETE` | `/api/categories/:id` | menghapus kategori |

---

### 🎨 3. Banner Promo API (`/api/banners`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/banners` | mengambil daftar banner aktif |
| `POST` | `/api/banners` | membuat banner promo baru |
| `PUT` | `/api/banners/:id` | mengedit banner promo |
| `DELETE` | `/api/banners/:id` | menghapus banner promo |

---

### 👤 4. Autentikasi & Pelanggan (`/api/auth` & `/api/customer`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `POST` | `/api/auth/login` | login administrator |
| `POST` | `/api/customer/register` | pendaftaran akun pelanggan baru |
| `POST` | `/api/customer/login` | login pelanggan |
| `GET` | `/api/customer/profile` | mengambil profil akun aktif |
| `PUT` | `/api/customer/profile` | memperbarui profil pelanggan |
| `GET` | `/api/admin/customers` | mengambil daftar seluruh pelanggan (Admin) |
| `GET` | `/api/admin/dashboard/stats` | mengambil statistik KPI penjualan & produk |

---

### 💬 5. WhatsApp & Live Chat API (`/api/wa` & `/api/chat`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/wa/qr` | mengambil QR Code login WhatsApp (WAHA) |
| `GET` | `/api/wa/status` | ngecek status koneksi WhatsApp (`WORKING` / `LOGGED_OUT`) |
| `POST` | `/api/wa/logout` | keluar dari sesi WhatsApp |
| `POST` | `/api/wa/webhook` | penampung webhook event pesan masuk dari WAHA |
| `GET` | `/api/chat/contacts` | mengambil daftar kontak percakapan |
| `GET` | `/api/chat/history/:jid` | mengambil riwayat percakapan per kontak |
| `POST` | `/api/chat/send` | mengirim pesan teks / media via WhatsApp |

---

### ⚡ 6. WebSockets Endpoint (`/ws`)

- **URL**: `ws://localhost:8080/ws`
- **Fungsi**: Membuka koneksi 2-arah real-time untuk penyiaran (broadcast) pesan masuk baru, perubahan status pesanan, dan pembaruan QR Code ke frontend secara instan.

---

## ⚙️ Pengaturan Environment Variables (`.env`)

Buat file `.env` di direktori `Bot_BE/` dengan konfigurasi berikut:

```env
PORT=8080
CLIENT_URL=http://localhost:5173
CORS_ALLOWED_ORIGINS=*

# Database Configuration
DB_DRIVER=sqlite
DB_SOURCE=data/bot.db

# WAHA WhatsApp Engine
WAHA_BASE_URL=http://localhost:3000
WAHA_API_KEY=your_secret_key
WAHA_SESSION=default

# JWT Authentication
JWT_SECRET=super_secret_jwt_key_12345
```

---

## 🧪 Cara Menjalankan Engine

### 1. Jalankan Secara Lokal (Go Native)
```bash
# Salin environment file
cp .env.example .env

# Unduh semua library dependencies
go mod tidy

# Jalankan server backend
go run cmd/main.go
```

### 2. Jalankan via Docker Compose
```bash
docker-compose up -d --build
```
