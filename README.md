# 🤖 Backend Microservice - Go Fiber E-Commerce API & WhatsApp Engine

Microservice API Gateway & Core Engine berbasis **Golang (Fiber v2)** & **GORM** untuk mengelola layanan E-Commerce, Autentikasi User/Admin, Pembayaran Midtrans, Cancel & Stock Restoration, WhatsApp Webhook Engine (WAHA), serta komunikasi real-time via WebSockets.

---

## 📋 Daftar Isi
- [🏗️ Arsitektur & Peran Microservice](#️-arsitektur--peran-microservice)
- [🌐 Tech Stack Backend](#-tech-stack-backend)
- [📁 Struktur Folder Backend (`internal/`)](#-struktur-folder-backend-internal)
- [🗄️ Model Data GORM (`internal/model/`)](#️-model-data-gorm-internalmodel)
- [📡 API Documentation Endpoints](#-api-documentation-endpoints)
- [🔒 Keamanan & Sistem Proteksi SQL Injection](#-keamanan--sistem-proteksi-sql-injection)
- [💳 Alur Pembayaran Idempotent & Midtrans Integration](#-alur-pembayaran-idempotent--midtrans-integration)
- [⚙️ Pengaturan Environment Variables (`.env`)](#️-pengaturan-environment-variables-env)
- [🧪 Cara Menjalankan Engine](#-cara-menjalankan-engine)

---

## 🏗️ Arsitektur & Peran Microservice

Microservice ini menangani seluruh proses bisnis kritis pada sistem:
1. 🛍️ **Core E-Commerce REST API**: Layanan CRUD Produk, Kategori, Banner Promo, Order History, dan Dashboard Stats.
2. 🔄 **Order Management & Stock Restoration**:
   - Pengelolaan status transaksi: `pending`, `paid`/`settlement`, `shipped`, `completed`, `canceled`.
   - **Fitur Pembatalan Pesanan (`PUT /api/orders/:id/cancel`)**: Mengembalikan jumlah stok barang (*stock restoration*) secara otomatis ke database dan mengirim sinyal WebSocket real-time.
3. 💳 **Midtrans Payment Integration**:
   - Pembuatan transaksi Snap / QRIS / Virtual Account (BCA, BNI, Mandiri, BRI).
   - Penampung callback webhook notifikasi otomatis dengan verifikasi `signature_key` & `gross_amount`.
4. 💬 **WhatsApp Gateway & Webhook Engine**:
   - Integrasi WAHA (WhatsApp HTTP API Provider) untuk otentikasi QR Code dan kontrol sesi.
   - Penampung Webhook pesan masuk real-time (`/api/wa/webhook`).
   - Live Chat & Manajemen Riwayat Pesan (`/api/chat/...`).
5. ⚡ **Real-Time WebSocket Hub (`wshub`)**: Penyiaran (*broadcasting*) event pembayaran (`payment_status_updated`), pembatalan (`order_canceled`), dan pesan baru secara instan ke frontend.
6. 🔐 **Role-Based Authentication & Multi-Layer Security**: System login terpisah untuk Admin dan Customer berbasis Bcrypt password hashing & parameterized SQL queries.

---

## 🌐 Tech Stack Backend

- 🐹 **Golang 1.20+** - Language Engine dengan konkurensi goroutine tinggi.
- ⚡ **Fiber v2 (`github.com/gofiber/fiber/v2`)** - Framework REST API tercepat untuk Go.
- 🗄️ **GORM (`gorm.io/gorm`)** - ORM Database PostgreSQL / MySQL / SQLite dengan `PrepareStmt: true`.
- 🔌 **Gorilla WebSockets (`github.com/gofiber/websocket/v2`)** - Komunikasi dua arah real-time.
- 💬 **WAHA (WhatsApp HTTP API)** - Provider sesi WhatsApp Web.
- 🐳 **Docker & Docker Compose** - Kontainerisasi produksi.

---

## 📁 Struktur Folder Backend (`internal/`)

```text
Bot_BE/
├── cmd/
│   └── app/
│       └── main.go                  # Entrypoint utama server
├── internal/
│   ├── config/
│   │   └── config.go                # Loader .env & variabel sistem
│   ├── provider/
│   │   ├── database.go              # Inisialisasi GORM, PrepareStmt, & Auto-Migration
│   │   ├── sheets.go                # Integrasi Google Sheets API
│   │   └── waha.go                  # Client integrasi WAHA WhatsApp Engine
│   ├── model/                       # Struct Model Database GORM
│   │   ├── product.go               # Model Produk
│   │   ├── category.go              # Model Kategori Produk
│   │   ├── banner.go                # Model Banner Promo
│   │   ├── customer.go              # Model Customer (Pelanggan)
│   │   ├── admin.go                 # Model Administrator
│   │   ├── order.go                 # Model Order & OrderItems
│   │   ├── config.go                # Model Konfigurasi Bot
│   │   ├── activity_log.go          # Model Audit Log Aktivitas
│   │   └── models.go                # Response DTO Global
│   ├── handler/                     # REST API Controllers
│   │   ├── product_handler.go       # Controller CRUD & Search Produk
│   │   ├── category_handler.go      # Controller CRUD Kategori & Auto-Seed
│   │   ├── banner_handler.go        # Controller CRUD Banner Promo
│   │   ├── auth_handler.go          # Controller Auth Login/Register Admin
│   │   ├── customer_handler.go      # Controller Auth Login/Register Customer
│   │   ├── order_handler.go         # Controller Order List, Detail, & Cancel
│   │   ├── payment_handler.go       # Controller Midtrans Checkout & Notification Callback
│   │   ├── message_handler.go       # Controller Webhook WA & Broadcast
│   │   ├── chat_handler.go          # Controller Live Chat History & Send
│   │   ├── config_handler.go        # Controller Spreadsheet Config
│   │   └── activity_handler.go      # Controller Log Aktivitas
│   ├── server/
│   │   └── server.go                # Router Fiber, CORS, Auth, & WS Upgrade
│   ├── service/
│   │   ├── order_service.go         # Logika Bisnis Order & Cancel Stock Restoration
│   │   ├── payment_service.go       # Logika Bisnis Midtrans & Notification Verification
│   │   ├── customer_service.go      # Logika Bisnis Customer
│   │   └── bot_service.go           # Logika Bisnis Bot WhatsApp
│   └── wshub/
│       └── hub.go                   # WebSocket Connection Hub & Broadcast Event
├── data/                            # File Storage (Database)
├── Dockerfile                       # Multi-Stage Production Dockerfile
├── docker-compose.yml               # Docker Compose Orchestration
└── go.mod                           # Go Module Manifest
```

---

## 🗄️ Model Data GORM (`internal/model/`)

### 1. Model Order (`order.go`)
```go
type Order struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
	OrderID         string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"order_id"`
	OrderNumber     string         `gorm:"type:varchar(100);default:''" json:"order_number"`
	IdempotencyKey  string         `gorm:"type:varchar(100);uniqueIndex;not null" json:"idempotency_key"`
	CustomerID      uint           `gorm:"index" json:"customer_id"`
	UserID          uint           `gorm:"index;default:0" json:"user_id"`
	CustomerName    string         `gorm:"type:varchar(255);default:''" json:"customer_name"`
	CustomerEmail   string         `gorm:"type:varchar(255);default:''" json:"customer_email"`
	CustomerPhone   string         `gorm:"type:varchar(50);default:''" json:"customer_phone"`
	ShippingAddress string         `gorm:"type:text;default:''" json:"shipping_address"`
	TotalAmount     float64        `gorm:"type:numeric(15,2);not null" json:"total_amount"`
	TotalPrice      float64        `gorm:"type:numeric(15,2);default:0" json:"total_price"`
	Status          string         `gorm:"type:varchar(50);default:'pending';index" json:"status"`
	SnapToken       string         `gorm:"type:text;default:''" json:"snap_token"`
	SnapRedirectURL string         `gorm:"type:text;default:''" json:"snap_redirect_url"`
	QRISURL         string         `gorm:"type:text;default:''" json:"qris_url,omitempty"`
	QRISString      string         `gorm:"type:text;default:''" json:"qris_string,omitempty"`
	VANumber        string         `gorm:"type:varchar(100);default:''" json:"va_number,omitempty"`
	VABank          string         `gorm:"type:varchar(50);default:''" json:"va_bank,omitempty"`
	PaymentType     string         `gorm:"type:varchar(50);default:''" json:"payment_type"`
	PaidAt          *time.Time     `json:"paid_at,omitempty"`
	OrderItems      []OrderItem    `gorm:"foreignKey:OrderID;references:OrderID" json:"items"`
}
```

---

## 📡 API Documentation Endpoints

### 🛍️ 1. Produk API (`/api/products`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `GET` | `/api/products` | Mengambil seluruh daftar produk |
| `GET` | `/api/products/search?q=macbook` | Pencarian cepat produk berdasarkan keyword/SKU |
| `GET` | `/api/products/:id` | Mengambil detail produk berdasarkan ID |
| `POST` | `/api/products` | Membuat produk baru |
| `POST` | `/api/products/bulk` | Import produk massal via CSV/JSON |
| `PUT` | `/api/products/:id` | Memperbarui data produk |
| `DELETE` | `/api/products/:id` | Menghapus produk |

---

### 💳 2. Payment & Order API (`/api/payment` & `/api/orders`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `POST` | `/api/payment/checkout` | Membuat transaksi Snap / QRIS / Virtual Account Midtrans |
| `POST` | `/api/payment/notification` | Webhook callback tidak langsung dari Midtrans |
| `GET` | `/api/orders` | Mengambil riwayat seluruh pesanan |
| `GET` | `/api/orders/:id` | Mengambil detail pesanan berdasarkan `order_id` |
| `PUT` | `/api/orders/:id/cancel` | Pembatalan pesanan pending & pengembalian stok barang (*Stock Restoration*) |
| `GET` | `/api/orders/:id/invoice` | Mengunduh file Invoice Pembayaran Resmi dalam format PDF (`application/pdf`) |

---

### 👤 3. Autentikasi & Customer (`/api/auth` & `/api/customer`)

| Method | Endpoint | Deskripsi |
| :--- | :--- | :--- |
| `POST` | `/api/auth/login` | Login administrator / admin (Sanitasi & Parameterized Query) |
| `POST` | `/api/customer/register` | Pendaftaran akun customer baru |
| `POST` | `/api/customer/login` | Login customer (Sanitasi & Parameterized Query) |
| `GET` | `/api/customer/profile` | Mengambil profil akun aktif |
| `PUT` | `/api/customer/profile` | Memperbarui profil customer |
| `GET` | `/api/admin/customers` | Mengambil daftar seluruh customer (Admin) |
| `GET` | `/api/admin/dashboard/stats` | Mengambil statistik KPI penjualan & produk |

---

### ⚡ 4. WebSockets Endpoint (`/ws`)

- **URL**: `ws://localhost:8080/ws` (atau `wss://` pada koneksi HTTPS)
- **Fungsi**: Penyiaran real-time event status pembayaran (`payment_status_updated`), pembatalan pesanan (`order_canceled`), serta update percakapan WhatsApp.

---

## 🔒 Keamanan & Sistem Proteksi SQL Injection

### 🛡️ 1. Parameterized Queries & Prepared Statements
Seluruh query database mengimplementasikan **Parameterized Placeholders (`?` / `$1`)** melalui GORM:
```go
h.DB.Where("email = ? OR phone = ?", username, username).First(&customer)
```
Di `internal/provider/database.go`, koneksi GORM diinisialisasi dengan `PrepareStmt: true`. Seluruh sintaks SQL dipre-kompilasi secara aman oleh server PostgreSQL sebelum memasukkan data pengguna. Karakter berbahaya seperti `' OR '1'='1` atau `; DROP TABLE` diperlakukan murni sebagai string data literal.

### 🛡️ 2. Input Sanitization & URL Traversal Protection
- Seluruh input email, phone, dan username disanitasi menggunakan `strings.TrimSpace()`.
- Pengisian parameter URL diproses melalui `encodeURIComponent()` untuk mencegah serangan *Path Traversal* atau *URL Injection*.

### 🛡️ 3. Bcrypt Password Hashing
Password disimpan menggunakan **Bcrypt Hashing Algorithm** dengan salt acak bawaan. Tidak ada password mentah yang disimpan di database atau ditampilkan pada log server.

---

## 💳 Alur Pembayaran Idempotent & Midtrans Integration

1. **At-Least-Once Webhook Protection**: Callback notifikasi Midtrans diproses di dalam transaksi database berbasis row-level locking (`clause.Locking{Strength: "UPDATE"}`).
2. **Amount & State Guard**: Backend memverifikasi `gross_amount` callback sesuai nilai transaksi database dan menolak rollback status jika transaksi sudah `paid`/`settlement`.
3. **Restorasi Stok Otomatis**: Jika pesanan dibatalkan (`PUT /api/orders/:id/cancel`) atau kadaluarsa (*expired*), stok barang dikembalikan secara otomatis ke tabel `products`.
4. **Kriptografi Idempotency Key**: Mencegah pembuatan transaksi ganda di Midtrans saat terjadi retries atau koneksi terputus.

---

## ⚙️ Pengaturan Environment Variables (`.env`)

Buat file `.env` di direktori `Bot_BE/`:

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
MIDTRANS_MERCHANT_ID=your_midtrans_merchant_id
MIDTRANS_CLIENT_KEY=your_midtrans_client_key
MIDTRANS_SERVER_KEY=your_midtrans_server_key
MIDTRANS_IS_PRODUCTION=false
```

---

## 🧪 Cara Menjalankan Engine

### 1. Jalankan secara Lokal (Go Native)
```bash
# Unduh library dependencies
go mod tidy

# Jalankan server backend
go run cmd/app/main.go
```

### 2. Jalankan via Docker Compose
```bash
docker-compose up -d --build
```
