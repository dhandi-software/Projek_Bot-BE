# 🤖 Backend Microservice - Projek Bot & E-Commerce

Layanan Backend microservice ini dibangun menggunakan **Golang** dan framework **Fiber v2**. Service ini bertanggung jawab menangani logika bisnis e-commerce (seperti **Debounced Product Search API**), integrasi WhatsApp Webhook via WAHA / Whatsmeow, autentikasi, serta komunikasi data ke PostgreSQL dan Google Sheets.

---

## 🏗️ Peran Microservice dalam Sistem

Microservice ini berperan sebagai API Gateway dan Core Engine:
- 🔍 **Debounced Product Search API**: Layanan REST API terpisah (`GET /api/products/search?q=...`) untuk melayani pencarian produk secara cepat dan efisien.
- 💬 **WhatsApp Engine Gateway**: Mengelola koneksi sesi WhatsApp, pengiriman pesan, penerimaan webhook real-time, dan penyimpanan riwayat obrolan.
- 📊 **Spreadsheet & Data Provider**: Mengelola konfigurasi ID Google Sheets dan aktivitas audit log.

---

## 🌐 Tech Stack

- 🐹 **Golang 1.22+** - Bahasa pemrograman backend performa tinggi.
- 🚀 **Fiber (v2)** - Framework HTTP server tercepat di ekosistem Go.
- 🗄️ **GORM & PostgreSQL** - Database relational dan ORM.
- 💬 **WAHA / Whatsmeow** - Library & engine integrasi WhatsApp.
- 🐳 **Docker & Docker Compose** - Kontainerisasi microservices.

---

## 📡 API Endpoints

### 🔍 Product Search Service
- `GET /api/products/search?q={query}`
  - Request: `http://localhost:8080/api/products/search?q=macbook`
  - Response:
    ```json
    {
      "query": "macbook",
      "total": 1,
      "results": [
        {
          "id": "1",
          "title": "MacBook Pro M3 Max 16-inch",
          "category": "Computer & Laptop",
          "price": "$2,499",
          "brand": "Apple",
          "image": "https://..."
        }
      ]
    }
    ```

### 💬 WhatsApp Service
- `GET /api/wa/qr` - Mengambil QR code login WhatsApp.
- `GET /api/wa/status` - Mengecek status sesi WA.
- `POST /api/chat/send` - Mengirim pesan WhatsApp.
- `POST /api/wa/webhook` - Endpoint penampung event webhook pesan masuk dari WAHA.

---

## 🧪 Cara Menjalankan

### Manual:
```bash
cp .env.example .env
go mod tidy
go run cmd/app/main.go
```

### Docker Compose:
```bash
docker-compose up -d --build
```
