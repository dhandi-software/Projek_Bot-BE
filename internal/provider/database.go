package provider

import (
	"fmt"
	"log"
	"time"

	"bot_be/internal/config"
	"bot_be/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDatabase menginisialisasi koneksi PostgreSQL dan menjalankan AutoMigrate
func InitDatabase(cfg *config.Config) (*gorm.DB, error) {
	// Auto create DB if not exists
	_ = createDBIfNotExists(cfg)

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBName, cfg.DBPort)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger:      logger.Default.LogMode(logger.Info),
		PrepareStmt: true,
	})

	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database postgres: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL")

	// Bersihkan NULL values & drop NOT NULL jika tabel products sudah ada dari versi sebelumnya
	db.Exec("UPDATE products SET title = name WHERE (title IS NULL OR title = '') AND name IS NOT NULL AND name != ''")
	db.Exec("UPDATE products SET title = '' WHERE title IS NULL")
	db.Exec("UPDATE products SET sku = '' WHERE sku IS NULL")
	db.Exec("UPDATE products SET category = '' WHERE category IS NULL")
	db.Exec("UPDATE products SET materials = '' WHERE materials IS NULL")
	db.Exec("UPDATE products SET brand = '' WHERE brand IS NULL")
	db.Exec("UPDATE products SET description = '' WHERE description IS NULL")
	db.Exec("UPDATE products SET image = '' WHERE image IS NULL")
	db.Exec("UPDATE products SET best_deal_started_at = NOW(), best_deal_expires_at = NOW() + INTERVAL '6 hours' WHERE is_best_deal = true AND (best_deal_expires_at > NOW() + INTERVAL '24 hours' OR best_deal_expires_at IS NULL)")

	db.Exec("ALTER TABLE products ALTER COLUMN name DROP NOT NULL")
	db.Exec("ALTER TABLE products ALTER COLUMN title DROP NOT NULL")
	db.Exec("ALTER TABLE products ALTER COLUMN sku DROP NOT NULL")
	db.Exec("ALTER TABLE products ALTER COLUMN category DROP NOT NULL")

	// Drop legacy foreign key constraints on order_items if present from earlier migrations
	db.Exec("ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_orders_order_items")
	db.Exec("ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_order_items_order")
	db.Exec("ALTER TABLE order_items DROP CONSTRAINT IF EXISTS fk_orders_items")

	// Ensure order_id and order_number columns in orders and order_items are VARCHAR(100)
	db.Exec("UPDATE orders SET order_number = order_id WHERE (order_number IS NULL OR order_number = '') AND order_id IS NOT NULL")
	db.Exec("UPDATE orders SET order_number = '' WHERE order_number IS NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN order_number DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN user_id DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN total_price DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN customer_id DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN customer_name DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN customer_email DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN customer_phone DROP NOT NULL")
	db.Exec("ALTER TABLE orders ALTER COLUMN shipping_address DROP NOT NULL")

	db.Exec("ALTER TABLE order_items ADD COLUMN IF NOT EXISTS order_id VARCHAR(100)")
	db.Exec("ALTER TABLE order_items ALTER COLUMN order_id TYPE VARCHAR(100) USING order_id::varchar")
	db.Exec("ALTER TABLE orders ALTER COLUMN order_id TYPE VARCHAR(100) USING order_id::varchar")
	db.Exec("ALTER TABLE orders ALTER COLUMN order_number TYPE VARCHAR(100) USING order_number::varchar")

	// Ensure order_items columns exist if table was created in earlier schema
	db.Exec("ALTER TABLE order_items ADD COLUMN IF NOT EXISTS title VARCHAR(255) DEFAULT ''")
	db.Exec("ALTER TABLE order_items ADD COLUMN IF NOT EXISTS product_id BIGINT DEFAULT 0")
	db.Exec("ALTER TABLE order_items ADD COLUMN IF NOT EXISTS quantity INT DEFAULT 1")
	db.Exec("ALTER TABLE order_items ADD COLUMN IF NOT EXISTS price NUMERIC(15,2) DEFAULT 0")
	db.Exec("UPDATE order_items SET title = name WHERE (title IS NULL OR title = '') AND name IS NOT NULL AND name != ''")
	db.Exec("UPDATE order_items SET title = product_name WHERE (title IS NULL OR title = '') AND product_name IS NOT NULL AND product_name != ''")

	// Jalankan Auto Migrate
	err = db.AutoMigrate(&model.Admin{}, &model.Customer{}, &model.AppConfig{}, &model.ActivityLog{}, &model.Product{}, &model.Banner{}, &model.Category{}, &model.Order{}, &model.OrderItem{})
	if err != nil {
		log.Println("Peringatan migrasi gabungan:", err)
		_ = db.AutoMigrate(&model.Admin{})
		_ = db.AutoMigrate(&model.Customer{})
		_ = db.AutoMigrate(&model.AppConfig{})
		_ = db.AutoMigrate(&model.ActivityLog{})
		_ = db.AutoMigrate(&model.Product{})
		_ = db.AutoMigrate(&model.Banner{})
		_ = db.AutoMigrate(&model.Category{})
		_ = db.AutoMigrate(&model.Order{})
		_ = db.AutoMigrate(&model.OrderItem{})
	}

	log.Println("Migrasi tabel berhasil")

	// Seed data awal jika belum ada
	seedAdmin(db)
	seedCustomer(db)
	seedCategories(db)
	seedProducts(db)
	seedBanners(db)

	return db, nil
}

func seedAdmin(db *gorm.DB) {
	var count int64
	db.Model(&model.Admin{}).Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		
		admin := model.Admin{
			Username: "admin",
			Password: string(hashedPassword),
		}
		db.Create(&admin)
		log.Println("Akun admin default berhasil dibuat (Username: admin, Password: admin123)")
	}
}

func seedCustomer(db *gorm.DB) {
	var count int64
	db.Model(&model.Customer{}).Count(&count)
	if count == 0 {
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("customer123"), bcrypt.DefaultCost)
		
		cust := model.Customer{
			Name:     "Customer Dhandi",
			Email:    "customer@gmail.com",
			Phone:    "081234567890",
			Password: string(hashedPassword),
			Address:  "Jakarta, Indonesia",
			Bio:      "Setia berbelanja",
		}
		db.Create(&cust)
		log.Println("Akun customer default berhasil dibuat (Email: customer@gmail.com, Password: customer123)")
	}
}

func seedProducts(db *gorm.DB) {
	var count int64
	db.Model(&model.Product{}).Count(&count)
	if count > 0 {
		return
	}

	now := time.Now()
	expires := now.Add(6 * time.Hour)

	products := []model.Product{
		{
			SKU:               "SKU-MACM3MAX",
			Title:             "MacBook Pro M3 Max 16-inch 36GB 1TB",
			Category:          "Computer & Laptop",
			Price:             42000000,
			DiscountPrice:     35999000,
			Stock:             15,
			LowStockThreshold: 5,
			Weight:            2.1,
			Materials:         "Recycled Aluminum, Liquid Retina XDR",
			Brand:             "Apple",
			ShortDescription:  "Laptop Pro paling powerful dengan chip Apple M3 Max.",
			Description:       "Performa super kencang untuk rendering 3D, coding berat, dan editing video 8K tanpa lag dengan layar Liquid Retina XDR 120Hz.",
			Image:             "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-SONYXM5",
			Title:             "Sony WH-1000XM5 Wireless Noise-Canceling Headphones",
			Category:          "Headphone",
			Price:             5999000,
			DiscountPrice:     4299000,
			Stock:             35,
			LowStockThreshold: 10,
			Weight:            0.25,
			Materials:         "Soft Fit Leather, Synthetic Polymer",
			Brand:             "Sony",
			ShortDescription:  "Headphone peredam bising terbaik dengan audio hi-res.",
			Description:       "Dilengkapi dengan 8 mikrofon dan dua prosesor untuk meredam kebisingan secara sempurna. Baterai tahan hingga 30 jam pemakaian.",
			Image:             "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-S24ULTRA",
			Title:             "Samsung Galaxy S24 Ultra 512GB Titanium Black",
			Category:          "Smartphone",
			Price:             21999000,
			DiscountPrice:     18499000,
			Stock:             25,
			LowStockThreshold: 5,
			Weight:            0.23,
			Materials:         "Titanium Frame, Gorilla Armor Glass",
			Brand:             "Samsung",
			ShortDescription:  "Flagship terbaik dengan fitur AI canggih dan kamera 200MP.",
			Description:       "Dilengkapi prosesor Snapdragon 8 Gen 3, layar Dynamic AMOLED 2X 120Hz, S Pen terintegrasi, serta teknologi Galaxy AI mutakhir.",
			Image:             "https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-PS5SLIM",
			Title:             "PlayStation 5 Slim Digital Edition 1TB SSD",
			Category:          "Gaming Console",
			Price:             8199000,
			DiscountPrice:     6999000,
			Stock:             20,
			LowStockThreshold: 5,
			Weight:            3.2,
			Materials:         "High Quality Polycarbonate",
			Brand:             "Sony",
			ShortDescription:  "Konsol game generasi terbaru dengan desain lebih ramping.",
			Description:       "Mainkan game generasi berikutnya dengan grafik 4K 120Hz, loading ultra cepat berkat SSD NVMe custom 1TB, dan kontroler DualSense Haptic Feedback.",
			Image:             "https://images.unsplash.com/photo-1606813907291-d86efa9b94db?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-DELLXPS15",
			Title:             "Dell XPS 15 OLED Touch Intel i9 32GB 1TB RTX 4070",
			Category:          "Computer & Laptop",
			Price:             32500000,
			DiscountPrice:     27999000,
			Stock:             12,
			LowStockThreshold: 3,
			Weight:            1.9,
			Materials:         "CNC Machined Aluminum, Carbon Fiber",
			Brand:             "Dell",
			ShortDescription:  "Laptop premium dengan layar OLED 3.5K Touchscreen.",
			Description:       "Kombinasi desain ultra-slim berbahan karbon fiber dan performa tinggi Intel Core i9 Gen 13 untuk kreator profesional.",
			Image:             "https://images.unsplash.com/photo-1593642632823-8f785ba67e45?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-IPADPRO129",
			Title:             "Apple iPad Pro 12.9-inch M2 Wi-Fi 256GB Space Gray",
			Category:          "Smartphone",
			Price:             19999000,
			DiscountPrice:     16799000,
			Stock:             18,
			LowStockThreshold: 4,
			Weight:            0.68,
			Materials:         "Recycled Aluminum, Liquid Retina XDR",
			Brand:             "Apple",
			ShortDescription:  "Tablet bertenaga chip M2 dengan layar Liquid Retina XDR.",
			Description:       "Mendukung Apple Pencil hover feature, Thunderbolt port, serta kamera ProRes 12MP untuk pengalaman kerja dan hiburan maksimal.",
			Image:             "https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-ROGZEPHYR",
			Title:             "ASUS ROG Zephyrus G16 OLED Intel Ultra 9 RTX 4080",
			Category:          "Computer & Laptop",
			Price:             36999000,
			DiscountPrice:     31499000,
			Stock:             8,
			LowStockThreshold: 2,
			Weight:            1.85,
			Materials:         "CNC Aluminum Alloy, OLED Panel",
			Brand:             "ASUS",
			ShortDescription:  "Laptop gaming tertipis dengan layar OLED ROG Nebula 240Hz.",
			Description:       "Ditenagai prosesor Intel Core Ultra 9 dengan akselerator AI dan kartu grafis NVIDIA RTX 4080 dalam bodi alumunium CNC premium.",
			Image:             "https://images.unsplash.com/photo-1603302576837-37561b2e2302?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-AIRPODSPRO",
			Title:             "Apple AirPods Pro 2nd Gen USB-C Active Noise Cancellation",
			Category:          "Headphone",
			Price:             4299000,
			DiscountPrice:     3299000,
			Stock:             50,
			LowStockThreshold: 10,
			Weight:            0.05,
			Materials:         "Glossy White Polycarbonate",
			Brand:             "Apple",
			ShortDescription:  "Earbuds nirkabel dengan chip H2 dan peredam bising 2x lebih efektif.",
			Description:       "Dilengkapi Audio Spasial Personalisasi, mode Transparansi Adaptif, dan MagSafe Case pengisian daya via kabel USB-C.",
			Image:             "https://images.unsplash.com/photo-1600294037681-c80b4cb5b434?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-MXMASTER3S",
			Title:             "Logitech MX Master 3S Wireless Performance Mouse",
			Category:          "Computer Accessories",
			Price:             1699000,
			DiscountPrice:     1299000,
			Stock:             45,
			LowStockThreshold: 10,
			Weight:            0.14,
			Materials:         "Tactile Rubber & Steel Scroll Wheel",
			Brand:             "Logitech",
			ShortDescription:  "Mouse ergononis produktivitas paling senyap dengan sensor 8K DPI.",
			Description:       "Scroll MagSpeed elektromagnetik dapat melakukan 1000 baris detik. Klik 90% lebih senyap dibanding versi pendahulunya.",
			Image:             "https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-SWITCHZELDA",
			Title:             "Nintendo Switch OLED Model Legend of Zelda Tears of the Kingdom",
			Category:          "Gaming Console",
			Price:             5499000,
			DiscountPrice:     4399000,
			Stock:             14,
			LowStockThreshold: 3,
			Weight:            0.42,
			Materials:         "Custom Textured Matte Shell",
			Brand:             "Nintendo",
			ShortDescription:  "Konsol edisi spesial Zelda dengan layar OLED 7 inci jernih.",
			Description:       "Nikmati warna cerah kontras tinggi di mana saja. Memiliki audio yang diperbarui, dock berport LAN, dan memori internal 64GB.",
			Image:             "https://images.unsplash.com/photo-1578303512597-81e6cc155b3e?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-BOSEQCULTRA",
			Title:             "Bose QuietComfort Ultra Headphones Immersive Audio",
			Category:          "Headphone",
			Price:             6499000,
			DiscountPrice:     5199000,
			Stock:             20,
			LowStockThreshold: 5,
			Weight:            0.25,
			Materials:         "Protein Leather & Cast Aluminum",
			Brand:             "Bose",
			ShortDescription:  "Headphone kelas dunia dengan teknologi Bose Immersive Audio.",
			Description:       "Menghadirkan suara spasial yang membuat musik terasa lebih nyata di hadapan Anda. Nyaman dipakai seharian dengan peredaman terdepan.",
			Image:             "https://images.unsplash.com/photo-1546435770-a3e426bf472b?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-IPHONE15PM",
			Title:             "Apple iPhone 15 Pro Max 256GB Natural Titanium",
			Category:          "Smartphone",
			Price:             24999000,
			DiscountPrice:     21999000,
			Stock:             16,
			LowStockThreshold: 4,
			Weight:            0.22,
			Materials:         "Grade 5 Titanium & Ceramic Shield",
			Brand:             "Apple",
			ShortDescription:  "iPhone paling ringan dan kuat dengan bodi Titanium kelas penerbangan.",
			Description:       "Bertenaga chip A17 Pro 3nm, kamera 48MP dengan zoom optik 5x telephoto, tombol Action yang dapat disesuaikan, dan port USB-C 3.",
			Image:             "https://images.unsplash.com/photo-1695048133142-1a20484d2569?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-KEYCHRONQ1",
			Title:             "Keychron Q1 Max QMK Wireless Custom Mechanical Keyboard",
			Category:          "Computer Accessories",
			Price:             3299000,
			DiscountPrice:     2699000,
			Stock:             25,
			LowStockThreshold: 5,
			Weight:            1.7,
			Materials:         "Full CNC Machined Aluminum Body",
			Brand:             "Keychron",
			ShortDescription:  "Keyboard mekanikal kustom 75% nirkabel berbahan aluminium solid.",
			Description:       "Mendukung koneksi 2.4GHz & Bluetooth 5.1, double-gasket design untuk sensasi mengetik empuk, dan keycaps PBT double-shot.",
			Image:             "https://images.unsplash.com/photo-1587829741301-dc798b83add3?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-LGOLED27",
			Title:             "LG UltraGear 27GR95QE 27-inch OLED 240Hz Gaming Monitor",
			Category:          "Computer Accessories",
			Price:             14999000,
			DiscountPrice:     11899000,
			Stock:             10,
			LowStockThreshold: 2,
			Weight:            5.0,
			Materials:         "Anti-glare OLED Panel, Hexagon Lighting Chassis",
			Brand:             "LG",
			ShortDescription:  "Monitor OLED gaming 240Hz dengan waktu respon ultra 0.03ms.",
			Description:       "Rasakan warna hitam sempurna dan kontras rasio 1.500.000:1 dengan teknologi NVIDIA G-SYNC & AMD FreeSync Premium Pro.",
			Image:             "https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-STEAMDECKOLED",
			Title:             "Valve Steam Deck OLED 512GB Handheld Gaming Console",
			Category:          "Gaming Console",
			Price:             10499000,
			DiscountPrice:     8799000,
			Stock:             15,
			LowStockThreshold: 4,
			Weight:            0.64,
			Materials:         "Matte Polycarbonate Handheld Body",
			Brand:             "Valve",
			ShortDescription:  "Perangkat gaming genggam dengan layar HDR OLED memukau.",
			Description:       "Mainkan ribuan game PC Steam Anda di mana saja. Layar OLED 90Hz, daya tahan baterai hingga 50% lebih lama, dan Wi-Fi 6E ultra cepat.",
			Image:             "https://images.unsplash.com/photo-1612287230202-1ff1d85d1bdf?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-JBLFLIP6",
			Title:             "JBL Flip 6 Waterproof Portable Bluetooth Speaker",
			Category:          "Headphone",
			Price:             2199000,
			DiscountPrice:     1599000,
			Stock:             60,
			LowStockThreshold: 15,
			Weight:            0.55,
			Materials:         "Durable Fabric & Rugged Rubber Housing",
			Brand:             "JBL",
			ShortDescription:  "Speaker portabel tahan air IP67 dengan suara Original JBL Pro.",
			Description:       "Sistem speaker 2-arah menghasilkan suara bass bertenaga dan nada tinggi jernih. Tahan air & debu dengan masa pakai baterai 12 jam.",
			Image:             "https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-WATCH6CLASSIC",
			Title:             "Samsung Galaxy Watch6 Classic 47mm Bluetooth Stainless Steel",
			Category:          "Smartphone",
			Price:             5999000,
			DiscountPrice:     4499000,
			Stock:             22,
			LowStockThreshold: 5,
			Weight:            0.06,
			Materials:         "Stainless Steel Body, Sapphire Crystal Glass",
			Brand:             "Samsung",
			ShortDescription:  "Smartwatch ikonik dengan bezel putar fisik khas Samsung.",
			Description:       "Layar 20% lebih besar, pelacak tidur tingkat lanjut, analisis komposisi tubuh BIA, serta kaca Sapphire Crystal yang tahan gores.",
			Image:             "https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-ANKER737",
			Title:             "Anker 737 Power Bank PowerCore 24,000mAh 140W",
			Category:          "Computer Accessories",
			Price:             2499000,
			DiscountPrice:     1799000,
			Stock:             30,
			LowStockThreshold: 8,
			Weight:            0.63,
			Materials:         "Fire-retardant Matte Casing & Smart Display",
			Brand:             "Anker",
			ShortDescription:  "Powerbank 140W ultra-powerful dengan smart digital display.",
			Description:       "Mampu mengisi daya MacBook Pro hingga 50% hanya dalam 28 menit. Dilengkapi layar digital cerdas yang menampilkan watt masukan & keluaran.",
			Image:             "https://images.unsplash.com/photo-1620799140408-edc6dcb6d633?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-RAZERDEATHADDER",
			Title:             "Razer DeathAdder V3 Pro Wireless Ergonomic Esports Mouse",
			Category:          "Computer Accessories",
			Price:             2499000,
			DiscountPrice:     1899000,
			Stock:             25,
			LowStockThreshold: 5,
			Weight:            0.06,
			Materials:         "Ultra-lightweight 63g Ergonomic Shell",
			Brand:             "Razer",
			ShortDescription:  "Mouse esports wireless paling ringan 63g pilihan atlet pro.",
			Description:       "Sensor Optik Razer Focus Pro 30K DPI, switch optik Gen-3 tanpa double-click, serta konektivitas HyperSpeed Wireless latensi super rendah.",
			Image:             "https://images.unsplash.com/photo-1629429408209-1f912961dbd8?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
		{
			SKU:               "SKU-PSVR2HORIZON",
			Title:             "Sony PlayStation VR2 Headset + Horizon Call of the Mountain Bundle",
			Category:          "Gaming Console",
			Price:             10999000,
			DiscountPrice:     8499000,
			Stock:             10,
			LowStockThreshold: 3,
			Weight:            0.56,
			Materials:         "HDR OLED Display, Sense Controllers",
			Brand:             "Sony",
			ShortDescription:  "Headset VR generasi terbaru untuk PS5 dengan visual 4K HDR.",
			Description:       "Rasakan dunia virtual secara nyata dengan layar 4K HDR, eye tracking cerdas, haptic feedback pada headset, dan kontroler PS VR2 Sense.",
			Image:             "https://images.unsplash.com/photo-1593508512255-86ab42a8e620?auto=format&fit=crop&w=600&q=80",
			Status:            "active",
			IsFeatured:        true,
			IsActive:          true,
			IsBestDeal:        true,
			BestDealStartedAt: &now,
			BestDealExpiresAt: &expires,
		},
	}

	for i := range products {
		var existing model.Product
		if err := db.Where("sku = ?", products[i].SKU).First(&existing).Error; err != nil {
			db.Create(&products[i])
		} else {
			db.Model(&existing).Updates(map[string]interface{}{
				"title":                products[i].Title,
				"category":             products[i].Category,
				"price":                products[i].Price,
				"discount_price":       products[i].DiscountPrice,
				"stock":                products[i].Stock,
				"low_stock_threshold":  products[i].LowStockThreshold,
				"materials":            products[i].Materials,
				"brand":                products[i].Brand,
				"short_description":    products[i].ShortDescription,
				"description":          products[i].Description,
				"image":                products[i].Image,
				"status":               products[i].Status,
				"is_featured":          products[i].IsFeatured,
				"is_active":            products[i].IsActive,
				"is_best_deal":         products[i].IsBestDeal,
				"best_deal_started_at": products[i].BestDealStartedAt,
				"best_deal_expires_at": products[i].BestDealExpiresAt,
			})
		}
	}
	log.Println("Seed data 20 produk Best Deals berhasil diperbarui/dibuat!")
}


func seedBanners(db *gorm.DB) {
	var count int64
	db.Model(&model.Banner{}).Count(&count)
	if count > 0 {
		return
	}

	banners := []model.Banner{
		{
			Title:      "Xbox Consoles",
			Subtitle:   "Save up to 50% on select Xbox games. Get 3 months of PC Game Pass for $2 USD.",
			Tagline:    "THE BEST PLACE TO PLAY",
			Image:      "/images/Image.png",
			PriceBadge: "$299",
			LinkUrl:    "/category-demo?category=gaming-console",
			ButtonText: "SHOP NOW",
			BgColor:    "#FFFFFF",
			IsActive:   true,
			SortOrder:  1,
		},
		{
			Title:      "MacBook Pro M3 Max",
			Subtitle:   "Mind-blowing graphics power for developers, creators, and power users.",
			Tagline:    "UNLEASH PRO POWER",
			Image:      "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=800&q=80",
			PriceBadge: "$2,499",
			LinkUrl:    "/product/2",
			ButtonText: "BUY NOW",
			BgColor:    "#18181B",
			IsActive:   true,
			SortOrder:  2,
		},
	}

	for i := range banners {
		db.Create(&banners[i])
	}
	log.Println("Seed data banner promo awal berhasil dibuat")
}

func seedCategories(db *gorm.DB) {
	var count int64
	db.Model(&model.Category{}).Count(&count)
	if count > 0 {
		return
	}

	initialCategories := []model.Category{
		{Name: "Computer & Laptop", Slug: "computer-laptop", Description: "Perangkat Komputer dan Laptop", IsActive: true},
		{Name: "Gaming Console", Slug: "gaming-console", Description: "Konsol Game dan Aksesoris", IsActive: true},
		{Name: "Smartphone", Slug: "smartphone", Description: "Ponsel Pintar dan Tablet", IsActive: true},
		{Name: "Headphone", Slug: "headphone", Description: "Headphone, Earphone, Audio", IsActive: true},
		{Name: "Computer Accessories", Slug: "computer-accessories", Description: "Aksesoris Komputer", IsActive: true},
		{Name: "Umum", Slug: "umum", Description: "Kategori Umum", IsActive: true},
	}

	for i := range initialCategories {
		db.Create(&initialCategories[i])
	}
	log.Println("Seed data kategori awal berhasil dibuat")
}

func createDBIfNotExists(cfg *config.Config) error {
	dsnRoot := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPassword, cfg.DBPort)

	dbRoot, err := gorm.Open(postgres.Open(dsnRoot), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	sqlDB, err := dbRoot.DB()
	if err == nil {
		defer sqlDB.Close()
	}

	var exists bool
	query := fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = '%s')", cfg.DBName)
	dbRoot.Raw(query).Scan(&exists)

	if !exists {
		log.Printf("Database '%s' belum ada di PostgreSQL, membuat database secara otomatis...\n", cfg.DBName)
		execQuery := fmt.Sprintf("CREATE DATABASE \"%s\"", cfg.DBName)
		if err := dbRoot.Exec(execQuery).Error; err != nil {
			return err
		}
	}
	return nil
}
