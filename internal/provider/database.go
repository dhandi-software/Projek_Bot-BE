package provider

import (
	"fmt"
	"log"

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
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database postgres: %w", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL")

	// Bersihkan NULL values & drop NOT NULL jika tabel products sudah ada dari versi sebelumnya
	db.Exec("UPDATE products SET title = '' WHERE title IS NULL")
	db.Exec("UPDATE products SET sku = '' WHERE sku IS NULL")
	db.Exec("UPDATE products SET category = '' WHERE category IS NULL")
	db.Exec("UPDATE products SET materials = '' WHERE materials IS NULL")
	db.Exec("UPDATE products SET brand = '' WHERE brand IS NULL")
	db.Exec("UPDATE products SET description = '' WHERE description IS NULL")
	db.Exec("UPDATE products SET image = '' WHERE image IS NULL")

	db.Exec("ALTER TABLE products ALTER COLUMN title DROP NOT NULL")
	db.Exec("ALTER TABLE products ALTER COLUMN sku DROP NOT NULL")
	db.Exec("ALTER TABLE products ALTER COLUMN category DROP NOT NULL")

	// Jalankan Auto Migrate
	err = db.AutoMigrate(&model.Admin{}, &model.Customer{}, &model.AppConfig{}, &model.ActivityLog{}, &model.Product{}, &model.Banner{})
	if err != nil {
		log.Println("Peringatan migrasi gabungan:", err)
		_ = db.AutoMigrate(&model.Admin{})
		_ = db.AutoMigrate(&model.Customer{})
		_ = db.AutoMigrate(&model.AppConfig{})
		_ = db.AutoMigrate(&model.ActivityLog{})
		_ = db.AutoMigrate(&model.Product{})
		_ = db.AutoMigrate(&model.Banner{})
	}

	log.Println("Migrasi tabel berhasil")

	// Seed data awal jika belum ada
	seedAdmin(db)
	seedCustomer(db)
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

	products := []model.Product{
		{
			SKU:         "XBOX-CON-01",
			Title:       "Xbox Series X Console 1TB",
			Category:    "Gaming Console",
			Price:       499.00,
			Stock:       25,
			Materials:   "Custom AMD Zen 2 CPU, RDNA 2 GPU, High-density alloy chassis",
			Brand:       "Microsoft",
			Description: "Next-generation gaming console with 12 teraflops of raw graphic processing power, 4K gaming at up to 120 FPS.",
			Image:       "/images/Image.png",
			IsFeatured:  true,
			IsActive:    true,
		},
		{
			SKU:         "MAC-M3-MAX",
			Title:       "MacBook Pro M3 Max 16-inch",
			Category:    "Computer & Laptop",
			Price:       2499.00,
			Stock:       10,
			Materials:   "Recycled Aluminum, Mini-LED Liquid Retina XDR",
			Brand:       "Apple",
			Description: "Unmatched performance with 16-core CPU, 40-core GPU, and up to 128GB unified memory.",
			Image:       "https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=600&q=80",
			IsFeatured:  true,
			IsActive:    true,
		},
		{
			SKU:         "DELL-XPS-15",
			Title:       "Dell XPS 15 OLED Touch Laptop",
			Category:    "Computer & Laptop",
			Price:       1899.00,
			Stock:       15,
			Materials:   "CNC Machined Aluminum, Carbon Fiber Palmrest",
			Brand:       "Dell",
			Description: "Stunning 3.5K OLED touchscreen display powered by Intel Core i9 and NVIDIA RTX 4070 graphics.",
			Image:       "https://images.unsplash.com/photo-1593642632823-8f785ba67e45?auto=format&fit=crop&w=600&q=80",
			IsFeatured:  true,
			IsActive:    true,
		},
		{
			SKU:         "SONY-WH1000XM5",
			Title:       "Sony WH-1000XM5 Wireless Noise-Canceling Headphones",
			Category:    "Headphone",
			Price:       399.00,
			Stock:       40,
			Materials:   "Synthetic leather earcups, recycled plastic body",
			Brand:       "Sony",
			Description: "Industry-leading noise canceling with two processors and eight microphones for unparalleled sound quality.",
			Image:       "https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=600&q=80",
			IsFeatured:  true,
			IsActive:    true,
		},
		{
			SKU:         "SAMSUNG-S24U",
			Title:       "Samsung Galaxy S24 Ultra 512GB",
			Category:    "Smartphone",
			Price:       1299.00,
			Stock:       30,
			Materials:   "Titanium frame, Corning Gorilla Armor glass",
			Brand:       "Samsung",
			Description: "Galaxy AI powered smartphone featuring 200MP camera, built-in S Pen, and Snapdragon 8 Gen 3.",
			Image:       "https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?auto=format&fit=crop&w=600&q=80",
			IsFeatured:  true,
			IsActive:    true,
		},
	}

	for i := range products {
		db.Create(&products[i])
	}
	log.Println("Seed data produk awal berhasil dibuat")
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
