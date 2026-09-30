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

### 💳 7. Midtrans Payment & Order API (`/api/payment` & `/api/orders`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `POST` | `/api/payment/checkout` | Membuat transaksi Snap / QRIS / Virtual Account Midtrans & Penampung Webhook Callback Notifikasi (Idempotency Key & Signature Verification) |
| `POST` | `/api/payment/notification` | (Backward Compatibility) Webhook callback alternatif dari Midtrans |
| `GET` | `/api/orders` | Mengambil seluruh riwayat transaksi order |
| `GET` | `/api/orders/:id` | Mengambil detail order berdasarkan `order_id` |
| `GET` | `/api/orders/:id/invoice` | Mengunduh file Invoice Pembayaran Resmi dalam format PDF (`application/pdf`) |

#### Contoh Payload Checkout Request (`POST /api/payment/checkout`)
```json
{
  "idempotency_key": "IDEM-1727438100-XYZ",
  "payment_method": "qris",
  "bank": "bca",
  "customer": {
    "customer_id": 1,
    "name": "Budi Santoso",
    "email": "budi@example.com",
    "phone": "081234567890",
    "address": "Jl. Sudirman No. 45, Jakarta"
  },
  "items": [
    {
      "product_id": 2,
      "quantity": 1,
      "title": "Sony PlayStation VR2 Headset",
      "price": 8499000,
      "image": "https://images.unsplash.com/photo-1622979135225-d2ba269bc1bd"
    }
  ]
}
```

#### Contoh Response Checkout Success (`200 OK`)
```json
{
  "message": "Transaksi berhasil dibuat",
  "data": {
    "order_id": "ORDER-1727438100123-IDEM-172",
    "snap_token": "a1b2c3d4-5678-90ab-cdef-1234567890ab",
    "snap_redirect_url": "https://app.sandbox.midtrans.com/snap/v2/vtweb/a1b2c3d4",
    "qris_url": "https://api.sandbox.midtrans.com/v2/qris/391038100123/qr-code",
    "qris_string": "00020101021226680016ID.CO.QRIS.WWW...",
    "va_number": "12345678901",
    "va_bank": "bca",
    "total_amount": 8499000,
    "status": "pending",
    "is_reused": false
  }
}
```

#### Contoh Response Detail Order (`GET /api/orders/:id`)
```json
{
  "status": "success",
  "data": {
    "id": 15,
    "order_id": "ORDER-1727438100123-IDEM-172",
    "customer_name": "Budi Santoso",
    "customer_email": "budi@example.com",
    "customer_phone": "081234567890",
    "customer_address": "Jl. Sudirman No. 45, Jakarta",
    "total_amount": 8499000,
    "payment_method": "qris",
    "status": "paid",
    "items": [
      {
        "id": 28,
        "order_id": "ORDER-1727438100123-IDEM-172",
        "product_id": 2,
        "title": "Sony PlayStation VR2 Headset",
        "quantity": 1,
        "price": 8499000,
        "subtotal": 8499000,
        "image": "https://images.unsplash.com/photo-1622979135225-d2ba269bc1bd",
        "image_url": "https://images.unsplash.com/photo-1622979135225-d2ba269bc1bd"
      }
    ],
    "created_at": "2026-09-27T22:00:00+07:00"
  }
}
```

---

## 🔒 Proteksi Keamanan & Idempotency Key

### 🛡️ Proteksi SQL Injection & Data Sanitization
Seluruh layer database menggunakan ORM **GORM** dengan **Parameterized Prepared Queries** (`db.Where("column = ?", value)`). Tidak ada string concatenation pada query database untuk mencegah serangan SQL Injection secara total. Log server disanitasi sehingga tidak membocorkan `ServerKey`, `SnapToken`, `signature_key`, atau kredensial sensitif lainnya.

### 🔄 Pembayaran Idempotent & Race Condition Protection
1. **At-Least-Once Webhook Protection**: Notifikasi Midtrans diproses di dalam transaksi database berbasis row-level locking (`clause.Locking{Strength: "UPDATE"}`).
2. **Amount & State Transition Guard**: Backend memverifikasi `gross_amount` callback sesuai nilai transaksi database dan menolak rollback status jika transaksi sudah `paid`/`settlement`.
3. **Validasi Produk Real-Time**: Seluruh harga dan ketersediaan stok diambil langsung dari database internal. Backend menolak jika harga atau stok tidak valid.
4. **Idempotency Key Request**: Jika client mengirim `idempotency_key` yang sama, backend langsung mengembalikan invoice/token yang sudah dibuat tanpa membuat transaksi ganda di Midtrans.

---

## ⚙️ Pengaturan Environment Variables (`.env`)

Buat file `.env` di direktori `Bot_BE/` (salin dari `.env.example`):

```env
PORT=8080
CLIENT_URL=http://localhost:5173
CORS_ALLOWED_ORIGINS=*

# Database Configuration
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=your_db_password
DB_NAME=Projek_Bot
DB_PORT=5432

# WAHA WhatsApp Engine
WAHA_API_URL=http://localhost:3001
N8N_WEBHOOK_URL=https://your-n8n-instance.cloud/webhook/your-webhook-id

# Midtrans Payment Gateway Configuration
# ⚠️ PENTING: Jangan commit file .env yang berisi kunci asli ke repositori git!
MIDTRANS_MERCHANT_ID=your_midtrans_merchant_id
MIDTRANS_CLIENT_KEY=your_midtrans_client_key
MIDTRANS_SERVER_KEY=your_midtrans_server_key
MIDTRANS_IS_PRODUCTION=false
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
go run cmd/app/main.go
```

### 2. Jalankan via Docker Compose
```bash
docker-compose up -d --build
```

