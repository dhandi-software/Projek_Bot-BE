package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"bot_be/internal/config"
	"bot_be/internal/model"
	"bot_be/internal/provider"
	"gorm.io/gorm"
)

type CategorySeed struct {
	Name        string
	Slug        string
	Description string
	Icon        string
	Brands      []string
	Images      []string
	ItemTypes   []string
}

func main() {
	log.Println("🌱 Memulai Seeder 300 Produk Database...")

	cfg := config.LoadConfig()
	db, err := provider.InitDatabase(cfg)
	if err != nil {
		log.Fatalf("❌ Gagal terhubung ke database: %v", err)
	}

	SeedCategoriesAndProducts(db)
	log.Println("✅ SEEDING 300 PRODUK DAN KATEGORI SELESAI DENGAN SUKSES!")
}

func SeedCategoriesAndProducts(db *gorm.DB) {
	categoriesData := []CategorySeed{
		{
			Name:        "Computer & Laptop",
			Slug:        "computer-laptop",
			Description: "Laptop, PC Desktop, Server, dan Ultrabook bertenaga tinggi untuk profesional.",
			Icon:        "Laptop",
			Brands:      []string{"Apple", "Dell", "ASUS", "Lenovo", "HP", "Acer", "MSI"},
			Images: []string{
				"https://images.unsplash.com/photo-1517336714731-489689fd1ca8?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1603302576837-37561b2e2302?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1593642632823-8f785ba67e45?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1588872657578-7efd1f1555ed?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1541807084-5c52b6b3adef?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Pro M3 Max 16-inch 36GB 1TB", "XPS 15 OLED Touch i9 32GB", "ROG Zephyrus G16 OLED RTX 4080",
				"ThinkPad X1 Carbon Gen 11 i7", "Spectre x360 14 Touch OLED", "Swift Go 14 OLED Intel Evo",
				"Stealth 16 AI Studio RTX 4070", "MacBook Air M3 15-inch 16GB", "Legion Pro 7i Gen 8 i9 RTX 4090",
				"ZenBook 14 OLED Ultra 7", "Predator Helios 18 Mini-LED", "IdeaPad Slim 5 Gen 8 Ryzen 7",
				"Vivobook Pro 15 OLED RTX 4060", "Envy x360 2-in-1 Touchscreen", "Alienware m18 R2 i9 RTX 4090",
				"Surface Laptop 5 15-inch i7", "TUF Gaming A15 Ryzen 9 RTX 4070", "MacBook Pro 14 M3 Pro 18GB",
				"Gram 17 Ultra Light i7 32GB", "Modern 14 Business Laptop i5", "ExpertBook B9 OLED Ultra Light",
				"Yoga Book 9i Dual Screen OLED", "OmniBook X AI PC Snapdragon X", "Galaxy Book4 Ultra i9 RTX 4070",
				"Pavilion Plus 14 OLED i7", "Blade 16 Dual Mini-LED RTX 4090", "LOQ 15 Gaming RTX 4050",
				"Latitude 7440 Ultralight i7", "MateBook X Pro 3.1K OLED", "Chromebook Plus CX34 Core i3",
			},
		},
		{
			Name:        "Smartphone",
			Slug:        "smartphone",
			Description: "Smartphone Android & iOS flagship, lipat, dan midrange terbaru.",
			Icon:        "Smartphone",
			Brands:      []string{"Apple", "Samsung", "Google", "Xiaomi", "OnePlus", "OPPO", "Vivo", "Realme"},
			Images: []string{
				"https://images.unsplash.com/photo-1610945265064-0e34e5519bbf?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1511707171634-5f897ff02aa9?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1598327105666-5b89351aff97?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1544244015-0df4b3ffc6b0?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1565849904461-04a58ad377e0?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Galaxy S24 Ultra 512GB Titanium", "iPhone 15 Pro Max 256GB Titanium", "Pixel 8 Pro 256GB Obsidian",
				"14 Ultra 512GB Leica Optics", "12 Pro 16GB 512GB Emerald", "Find N3 Fold 512GB Hasselblad",
				"X100 Pro 512GB Zeiss Camera", "GT 5 Pro 512GB Snapdragon 8", "Galaxy Z Fold5 512GB Phantom",
				"iPhone 15 Plus 128GB Pink", "Galaxy Z Flip5 256GB Mint", "Pixel 8 128GB Hazel",
				"Poco F6 Pro 512GB 120W Charge", "Redmagic 9 Pro 16GB Gaming Phone", "Nothing Phone (2) 256GB Glyph",
				"Galaxy A55 5G 256GB Awesome Navy", "Redmi Note 13 Pro+ 512GB 200MP", "iPhone 14 Pro Max 256GB Deep Purple",
				"Edge 50 Ultra 512GB Vegan Leather", "Zenfone 11 Ultra 512GB AI Phone", "Galaxy S23 FE 256GB Mint",
				"iPhone 13 128GB Starlight", "Find X7 Ultra Dual Periscope 512GB", "Reno11 Pro 5G 512GB Portrait Master",
				"iQOO 12 512GB Snapdragon 8 Gen 3", "Honor Magic6 Pro 512GB Falcon Camera", "Nord 4 512GB Metal Unibody",
				"Galaxy A35 5G 128GB Awesome Iceblue", "C67 256GB 108MP Camera", "Camo Edition 256GB Gaming Phone",
			},
		},
		{
			Name:        "Headphone",
			Slug:        "headphone",
			Description: "Headphone wireless noise canceling, TWS earbuds, dan studio monitor.",
			Icon:        "Headphones",
			Brands:      []string{"Sony", "Bose", "Apple", "Sennheiser", "JBL", "Audio-Technica", "Anker Soundcore"},
			Images: []string{
				"https://images.unsplash.com/photo-1505740420928-5e560c06d30e?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1546435770-a3e426bf472b?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1590658268037-6bf12165a8df?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1600294037681-c80b4cb5b434?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1583394838336-acd977736f90?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"WH-1000XM5 Noise Canceling ANC", "QuietComfort Ultra Wireless Headphones", "AirPods Pro 2nd Gen USB-C MagSafe",
				"MOMENTUM 4 Wireless Hi-Res Audio", "AirPods Max Space Gray Smart Case", "ATH-M50xBT2 Professional Studio",
				"Liberty 4 NC Noise Canceling TWS", "Tune 770NC Wireless Over-Ear", "WF-1000XM5 True Wireless Earbuds",
				"QuietComfort Earbuds II ANC", "Tour One M2 Spatial Sound Headphones", "HD 660S2 Open-Back Audiophile Headphones",
				"Space Q45 Adaptive ANC Headphones", "Tune 230NC TWS Pure Bass", "ULT WEAR Wireless ANC Heavy Bass",
				"ATH-SQ1TW Compact True Wireless", "MOMENTUM True Wireless 4 Snapdragon", "Live Pro 2 TWS 40H Playtime",
				"Space A40 Auto-Adjustable ANC", "LinkBuds S Ultra Light Wireless TWS", "Life Q30 Hybrid ANC Headphones",
				"QuietComfort 45 Classic Headphones", "AirPods 3rd Gen Spatial Audio", "Wave Flex TWS Ergonomic Fit",
				"ATH-M20xBT Wireless Studio Monitor", "Soundform Pulse Noise Canceling Earbuds", "Live 670NC On-Ear Wireless ANC",
				"Space One Adaptive Noise Canceling", "WF-C700N Compact Noise Canceling", "Studio Pro Wireless Spatial Audio",
			},
		},
		{
			Name:        "Gaming Console",
			Slug:        "gaming-console",
			Description: "Konsol game rumah, handheld PC, VR headset, dan kontroler nirkabel.",
			Icon:        "Gamepad2",
			Brands:      []string{"Sony", "Microsoft", "Nintendo", "Valve", "ASUS", "Lenovo"},
			Images: []string{
				"https://images.unsplash.com/photo-1606813907291-d86efa9b94db?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1578303512597-81e6cc155b3e?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1612287230202-1ff1d85d1bdf?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1622979135225-d2ba269bc1bd?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1593508512255-86ab42a8e620?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"PlayStation 5 Slim Digital 1TB SSD", "Xbox Series X 1TB 4K Console", "Nintendo Switch OLED Zelda Edition",
				"Steam Deck OLED 512GB Handheld PC", "ROG Ally X Ryzen Z1 Extreme 1TB", "Legion Go 8.8-inch QHD+ 144Hz Handheld",
				"PlayStation VR2 Headset 4K HDR", "DualSense Edge Wireless Controller Pro", "Xbox Elite Wireless Controller Series 2",
				"PlayStation Portal Remote Player PS5", "Nintendo Switch Lite Turquoise Portable", "PS5 Disc Drive Edition 1TB Marvel Spider-Man",
				"Xbox Series S 1TB Carbon Black", "PlayStation 5 DualSense Wireless Charging Dock", "8BitDo Ultimate Wireless Controller Dock",
				"Meta Quest 3 128GB VR Headset", "Backbone One Mobile Gaming Controller USB-C", "Razer Kishi V2 Pro Mobile Controller",
				"PSVR2 Horizon Call of the Mountain Bundle", "Nintendo Switch Pro Controller Wireless", "Xbox Wireless Controller Velocity Green",
				"Ayaneo 2S Ryzen 7840U 32GB 2TB", "GPD WIN 4 2024 Handheld Gaming PC", "Miyoo Mini Plus Portable Retro Console",
				"Anbernic RG35XX H Metal Shell Retro", "Sega Genesis Mini 2 Console Classic", "Atari 2600+ HDMI Retro Gaming System",
				"PlayStation 4 Pro 1TB Jet Black Classic", "Xbox Game Pass Ultimate 12-Month Card", "SN30 Pro Bluetooth Gamepad Retro",
			},
		},
		{
			Name:        "Computer Accessories",
			Slug:        "computer-accessories",
			Description: "Mouse ergonomic, keyboard mekanikal, monitor high-refresh rate, & hub.",
			Icon:        "Keyboard",
			Brands:      []string{"Logitech", "Razer", "Keychron", "Corsair", "LG", "Samsung", "SteelSeries", "Anker"},
			Images: []string{
				"https://images.unsplash.com/photo-1615663245857-ac93bb7c39e7?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1587829741301-dc798b83add3?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1527443224154-c4a3942d3acf?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1629429408209-1f912961dbd8?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1609592424074-2790757754d9?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"MX Master 3S Ergonomic Silent Mouse", "Q1 Max QMK Wireless Mechanical Keyboard", "UltraGear 27-inch OLED 240Hz Gaming Monitor",
				"DeathAdder V3 Pro Wireless 63g Mouse", "Odyssey G9 49-inch Curved OLED 240Hz", "Apex Pro TKL Wireless Mechanical Keyboard",
				"MX Keys S Wireless Illuminated Keyboard", "Viper V3 Pro Ultra-lightweight Wireless Mouse", "Odyssey G7 32-inch 4K 144Hz Gaming Monitor",
				"737 Power Bank 24,000mAh 140W Smart Display", "K2 Pro Wireless Mechanical Keyboard RGB", "Superlight 2 Wireless Gaming Mouse 60g",
				"UltraFine 32-inch 4K Ergo Monitor USB-C", "Huntsman V3 Pro TKL Analog Optical Keyboard", "Arctis Nova Pro Wireless Headset Dual DAC",
				"Stream Deck MK.2 15 Customizable Keys", "PowerExpand 13-in-1 USB-C Docking Station", "G502 X PLUS LIGHTSPEED Wireless RGB Mouse",
				"MagSpeed Wireless Vertical Ergonomic Mouse", "K3 Pro Ultra-Slim Wireless Mechanical Keyboard", "C920x HD Pro Webcam 1080p Auto-Focus",
				"SmartDock 10-in-1 Dual 4K HDMI Hub", "Gigantus V2 XXL Soft Gaming Mouse Pad", "Lift Vertical Ergonomic Mouse Bluetooth",
				"G915 TKL Wireless RGB Mechanical Keyboard", "Brio 4K Ultra HD Webcam HDR RightLight", "G Pro X Wireless Gaming Headset 7.1",
				"PowerWave 15W Fast Wireless Charger Stand", "K70 MAX RGB Magnetic-Mechanical Keyboard", "Stratus Duo Wireless Gaming Controller PC",
			},
		},
		{
			Name:        "Smart Home & Automation",
			Slug:        "smart-home",
			Description: "Perangkat rumah pintar, kamera pengawas IP, lampu pintar, dan IoT hub.",
			Icon:        "Home",
			Brands:      []string{"Xiaomi", "TP-Link Tapo", "Philips Hue", "Google Nest", "Bardi", "Ezviz", "Aqara"},
			Images: []string{
				"https://images.unsplash.com/photo-1558002038-1055907df827?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1507646221600-784187abf738?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1543512214-318c7553f230?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Smart Security Camera C300 2K 360", "Smart Plug Wi-Fi Energy Monitor Energy Meter", "Smart Bulb RGB Color E27 Wi-Fi",
				"Nest Hub 2nd Gen Smart Display Speaker", "Smart Door Lock Fingerprint Card Passcode", "C6N 1080p Smart Wi-Fi Pan Tilt Camera",
				"Smart Hub M2 Zigbee 3.0 Home Automation", "Smart Air Purifier 4 Pro HEPA Filter", "Hue White & Color Ambiance Starter Kit",
				"Smart Robot Vacuum Cleaner Mop 2 Ultra", "Tapo P110 Mini Smart Wi-Fi Socket", "Smart Doorbell Video Camera Night Vision",
				"Smart Light Strip Plus RGB 2 Meter", "Smart Curtain Driver Wireless Motorized", "Nest Doorbell Battery Powered Wireless",
				"Tapo C210 Pan/Tilt Smart Security Camera", "Smart Thermostat E Temperature Control", "Smart Water Leak Sensor Alarm Alert",
				"Smart IR Remote Universal Controller", "Smart Wall Switch No Neutral Touch Panel", "Smart Motion Sensor Zigbee Detector",
				"Smart Smoke Detector Siren Alarm System", "Tapo L530E Smart Wi-Fi Multicolor Bulb", "Smart Feeder Automatic Pet Food Dispenser",
				"Smart Table Lamp Dimmable Atmosphere", "Smart Garage Door Opener Controller Kit", "Smart Outdoor Floodlight Camera 2K ANC",
				"Smart Dehumidifier Air Dryer 20L", "Smart Aroma Diffuser RGB Humidifier", "Smart Radiator Thermostat Valve Zigbee",
			},
		},
		{
			Name:        "Wearables & Smartwatch",
			Slug:        "wearables",
			Description: "Smartwatch kesehatan, fitness tracker, dan jam tangan pintar olahraga.",
			Icon:        "Watch",
			Brands:      []string{"Apple", "Samsung", "Garmin", "Fitbit", "Huawei", "Amazfit"},
			Images: []string{
				"https://images.unsplash.com/photo-1523275335684-37898b6baf30?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1579586337278-3befd40fd17a?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1508685096489-7aacd43bd3b1?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Watch Ultra 2 GPS + Cellular 49mm Titanium", "Galaxy Watch6 Classic 47mm Bluetooth Stainless", "Forerunner 965 AMOLED Premium Running",
				"Watch Series 9 GPS 45mm Aluminum", "Galaxy Watch Fit3 AMOLED Battery 13 Days", "Charge 6 Advanced Health & Fitness Tracker",
				"Watch GT 4 46mm Brown Leather Strap", "Cheetah Pro Premium Lightweight Running Watch", "Fenix 7X Pro Sapphire Solar Multisport",
				"Venu 3 AMOLED GPS Smartwatch Voice", "Watch SE 2nd Gen GPS 44mm Midnight", "Band 8 Ultra Slim Amoled Tracker",
				"Balance AI Fitness Tracker AMOLED", "Instinct 2X Solar Tactical Edition GPS", "Sense 2 Advanced Health Smartwatch",
				"Galaxy Watch6 44mm Bluetooth Sapphire", "Watch Fit 3 Square Display Metallic", "T-Rex 2 Rugged Outdoor GPS Smartwatch",
				"Approach S70 Golf GPS Smartwatch", "Watch 4 Pro Titanium Sapphire Glass", "Bip 5 Big Screen Smartwatch GPS Call",
				"Lily 2 Stylish Small Smartwatch", "Galaxy Watch FE 40mm Bluetooth Ceramic", "Band 9 Heart Rate SpO2 Tracker",
				"Active Edge Rugged Sport Smartwatch", "Forerunner 265 Music AMOLED Running", "Inspire 3 Fitness Tracker 10 Days Battery",
				"Watch Ultimate Expedition Steel Case", "GTR 4 HD AMOLED Dual Band GPS", "Vivomove Trend Hybrid Smartwatch Wireless",
			},
		},
		{
			Name:        "Camera & Photography",
			Slug:        "camera-photography",
			Description: "Kamera mirrorless, dslr, action cam 4K, dan gimbal stabilizer.",
			Icon:        "Camera",
			Brands:      []string{"Sony", "Canon", "Fujifilm", "DJI", "GoPro", "Nikon"},
			Images: []string{
				"https://images.unsplash.com/photo-1516035069371-29a1b244cc32?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1526170375885-4d8ecf77b99f?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1502920917128-1aa500764cbd?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Alpha a7 IV Mirrorless Camera Body", "EOS R6 Mark II Full Frame Mirrorless", "X-T5 Mirrorless Digital Camera Body",
				"Osmo Pocket 3 Gimbal Camera 4K 120fps", "HERO12 Black Action Camera 5.3K Video", "Z6 II Full Frame Mirrorless Camera",
				"Alpha a6700 APS-C Mirrorless Camera", "EOS R10 4K Mirrorless Vlogging Camera", "X100VI Digital Compact Camera 40MP",
				"Osmo Action 4 Adventure Combo Camera", "HERO11 Black Mini Compact Action Cam", "Z f Full Frame Retro Mirrorless Camera",
				"Alpha a7R V 61MP High Resolution Camera", "EOS R8 Compact Full Frame Mirrorless", "Mini EVO Hybrid Instant Camera Printer",
				"Ronin RS 3 Pro 3-Axis Gimbal Stabilizer", "Osmo Mobile 6 Smartphone Gimbal", "Z v10 Vlogging Digital Camera 4K",
				"Alpha ZV-E1 Full Frame Vlogging Camera", "EOS R50 Vlogging Mirrorless Kit", "Instax Mini 12 Instant Camera Pastel",
				"PowerShot V10 Pocket Vlogging Camera", "HERO10 Black 4K Waterproof Action Cam", "Alpha a7C II Compact Full Frame Camera",
				"X-S20 Mirrorless Camera Body 6K Video", "Osmo Pocket 2 Handheld 4K Gimbal Cam", "EOS R7 High Speed APS-C Camera",
				"Z 30 Mirrorless Vlogging Camera Kit", "Ronin SC 2 Lightweight 3-Axis Gimbal", "Pocket Cinema Camera 6K Pro Super35",
			},
		},
		{
			Name:        "TV & Home Entertainment",
			Slug:        "tv-entertainment",
			Description: "Smart TV OLED/QLED 4K, soundbar Dolby Atmos, dan proyektor bioskop.",
			Icon:        "Tv",
			Brands:      []string{"Samsung", "LG", "Sony", "TCL", "Hisense", "Xiaomi"},
			Images: []string{
				"https://images.unsplash.com/photo-1593359677879-a4bb92f829d1?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1540555700478-4be289fbecef?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1461151304267-38535e780c79?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Neo QLED 65-inch 4K Smart TV QN90C", "OLED Evo 55-inch 4K Smart TV C3 Series", "Bravia XR 65-inch OLED 4K Google TV A80L",
				"QLED 55-inch 4K Google TV 55C645", "U8K Mini-LED 65-inch 4K ULED Google TV", "TV A Pro 55-inch 4K UHD Google TV",
				"The Frame 55-inch 4K Smart Art TV QLED", "OLED G3 65-inch 4K Gallery Edition TV", "Bravia X90L 55-inch Full Array LED 4K",
				"HT-A7000 7.1.2ch Dolby Atmos Soundbar", "HW-Q990C 11.1.4ch Wireless Dolby Atmos Soundbar", "Soundbar S80QY 3.1.3ch Triple Height Channel",
				"Freestyle Gen 2 Portable Smart Projector", "Laser Projector 4K Ultra Short Throw 150-inch", "Mi Smart Projector 2 Full HD Compact",
				"Crystal UHD 50-inch 4K Smart TV CU8000", "UR80 4K Smart UHD TV 43-inch AI ThinQ", "Google TV 43-inch Full HD HDR Smart TV",
				"QLED 43-inch 4K Smart TV Q60C", "OLED Flex 42-inch Bendable Gaming TV", "Bravia X80L 43-inch 4K HDR Google TV",
				"Soundbar HW-B550 2.1ch Subwoofer Bass", "Soundbar Cinema SB190 2.1ch Dolby Atmos", "C845 Mini-LED 55-inch 144Hz Gaming TV",
				"Laser TV 100-inch 4K Trichroma Ultra Short", "Smart Projector Horizon Pro 4K True 2200 ANSI", "Neo QLED 85-inch 8K Smart TV QN900C",
				"OLED 77-inch 4K Smart TV G3 Series", "Bravia XR 85-inch Mini-LED 4K X95L", "Soundbar HT-S20R 5.1ch Real Surround",
			},
		},
		{
			Name:        "Audio & Speakers",
			Slug:        "audio-speakers",
			Description: "Speaker Bluetooth nirkabel, sistem audio rumah, dan mic podcast.",
			Icon:        "Speaker",
			Brands:      []string{"JBL", "Bose", "Marshall", "Sonos", "Harman Kardon", "Shure", "Rode"},
			Images: []string{
				"https://images.unsplash.com/photo-1608043152269-423dbba4e7e1?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1545454675-3531b543be5d?auto=format&fit=crop&w=600&q=80",
				"https://images.unsplash.com/photo-1511379938547-c1f69419868d?auto=format&fit=crop&w=600&q=80",
			},
			ItemTypes: []string{
				"Charge 5 Waterproof Portable Speaker", "SoundLink Flex Bluetooth Portable Speaker", "Stanmore III Wireless Bluetooth Speaker",
				"Era 300 Smart Speaker Dolby Atmos", "Aura Studio 4 Bluetooth Desktop Speaker", "SM7B Cardioid Dynamic Vocal Microphone",
				"Rodecaster Pro II Integrated Audio Production Studio", "PartyBox 310 Portable Party Speaker 240W", "SoundLink Revolve+ II Bluetooth Speaker",
				"Emberton II Compact Portable Speaker 30H", "Move 2 Battery Powered Smart Speaker", "Onyx Studio 8 Wireless Bluetooth Speaker",
				"MV7 USB/XLR Podcast Dynamic Microphone", "Wireless GO II Dual Channel Wireless Mic System", "Boombox 3 Portable Bluetooth Speaker Bass",
				"Home Speaker 500 Stereo Sound Voice Control", "Acton III Wireless Bluetooth Home Speaker", "Roam 2 Ultra Portable Waterproof Smart Speaker",
				"Go 3 Mini Portable Waterproof Bluetooth Speaker", "Micro Wireless Mic Compact Lapel System", "Clip 4 Portable Waterproof Speaker Carabiner",
				"SoundLink Micro Rugged Bluetooth Speaker", "Middleton Heavy Portable Bluetooth Speaker", "Five High-Fidelity Smart Speaker",
				"Citation 200 Portable Smart Speaker HD", "PodMic Dynamic Cardioid Podcast Microphone", "NT-USB Mini Studio Quality USB Microphone",
				"Extreme 3 Portable Bluetooth Speaker Strap", "SoundLink Mini II Special Edition Speaker", "Woburn III Bluetooth Wireless Home Speaker",
			},
		},
	}

	rand.Seed(time.Now().UnixNano())

	// 1. Seed Categories
	log.Println("📁 Menyiapkan 10 Kategori Utama...")
	for _, c := range categoriesData {
		var cat model.Category
		if err := db.Where("slug = ?", c.Slug).First(&cat).Error; err != nil {
			newCat := model.Category{
				Name:        c.Name,
				Slug:        c.Slug,
				Description: c.Description,
				Icon:        c.Icon,
				IsActive:    true,
			}
			db.Create(&newCat)
			log.Printf("  + Kategori dibuat: %s\n", c.Name)
		} else {
			db.Model(&cat).Updates(map[string]interface{}{
				"name":        c.Name,
				"description": c.Description,
				"icon":        c.Icon,
				"is_active":   true,
			})
		}
	}

	// 2. Generate 300 Products (30 per category)
	log.Println("📦 Menyiapkan 300 Produk Database...")

	var totalProductsCreated int
	var totalProductsUpdated int

	now := time.Now()
	bestDealExpires := now.Add(12 * time.Hour)

	productIndex := 1

	for _, catData := range categoriesData {
		for i, itemType := range catData.ItemTypes {
			brand := catData.Brands[i%len(catData.Brands)]
			img := catData.Images[i%len(catData.Images)]
			title := fmt.Sprintf("%s %s", brand, itemType)
			sku := fmt.Sprintf("SKU-%s-%03d", catData.Slug[:3], productIndex)

			// Generate prices in IDR (Rp 45.000 - Rp 38.000.000)
			basePrice := float64((rand.Intn(350) + 5) * 50000) // Rp 250.000 - Rp 17.500.000 approx
			if catData.Slug == "computer-laptop" || catData.Slug == "camera-photography" || catData.Slug == "tv-entertainment" {
				basePrice = float64((rand.Intn(400) + 50) * 100000) // Rp 5.000.000 - Rp 45.000.000
			} else if catData.Slug == "smartphone" || catData.Slug == "gaming-console" {
				basePrice = float64((rand.Intn(250) + 20) * 100000) // Rp 2.000.000 - Rp 27.000.000
			}

			// Discount price
			discountPrice := basePrice * 0.85
			if i%3 == 0 {
				discountPrice = basePrice * 0.75
			} else if i%5 == 0 {
				discountPrice = basePrice * 0.90
			}

			// Best Deal & Featured flags
			isFeatured := (productIndex%6 == 0 || productIndex <= 20)
			isBestDeal := (productIndex%10 == 0 || productIndex <= 15)

			var dealStarted *time.Time
			var dealExpires *time.Time

			if isBestDeal {
				dealStarted = &now
				dealExpires = &bestDealExpires
			}

			stock := rand.Intn(80) + 10
			weight := float64(rand.Intn(50)+1) / 10.0

			shortDesc := fmt.Sprintf("%s produk original garansi resmi 1 tahun. Stok ready terjamin.", title)
			desc := fmt.Sprintf("Performa luar biasa dari %s %s. Dibuat dengan material premium tahan lama, fitur canggih terbaru, serta dukungan garansi resmi distributor Indonesia.", brand, itemType)

			p := model.Product{
				SKU:               sku,
				Title:             title,
				Category:          catData.Name,
				Price:             basePrice,
				DiscountPrice:     discountPrice,
				Stock:             stock,
				LowStockThreshold: 5,
				Weight:            weight,
				Materials:         "Aluminum Alloy & High-grade Polycarbonate",
				Brand:             brand,
				ShortDescription:  shortDesc,
				Description:       desc,
				Image:             img,
				Status:            "active",
				IsFeatured:        isFeatured,
				IsActive:          true,
				IsBestDeal:        isBestDeal,
				BestDealStartedAt: dealStarted,
				BestDealExpiresAt: dealExpires,
				Features:          "[\"100% Produk Original Garansi Resmi\",\"Pengiriman Cepat & Aman All Indonesia\",\"Dukungan Layanan Pelanggan 24/7\",\"Jaminan Uang Kembali Jika Barang Rusak\"]",
				Colors:            "[{\"id\":\"black\",\"name\":\"Black / Dark\",\"hex\":\"#18181B\"},{\"id\":\"silver\",\"name\":\"Silver / Metal\",\"hex\":\"#E4E4E7\"}]",
				ShippingInfo:      "\"Dikirim dari Gudang Utama Jakarta. Estimasi 1 - 3 hari kerja.\"",
				AdditionalInfo:    fmt.Sprintf("{\"Brand\":\"%s\",\"Garansi\":\"1 Tahun Garansi Resmi\",\"Model\":\"%s\"}", brand, sku),
				Specifications:    fmt.Sprintf("{\"Merek\":\"%s\",\"SKU\":\"%s\",\"Kondisi\":\"Baru 100%% Original\",\"Garansi\":\"1 Tahun Official Warranty\"}", brand, sku),
			}

			var existing model.Product
			if err := db.Where("sku = ?", sku).First(&existing).Error; err != nil {
				db.Create(&p)
				totalProductsCreated++
			} else {
				db.Model(&existing).Updates(map[string]interface{}{
					"title":                p.Title,
					"category":             p.Category,
					"price":                p.Price,
					"discount_price":       p.DiscountPrice,
					"stock":                p.Stock,
					"brand":                p.Brand,
					"short_description":    p.ShortDescription,
					"description":          p.Description,
					"image":                p.Image,
					"status":               "active",
					"is_featured":          p.IsFeatured,
					"is_active":            true,
					"is_best_deal":         p.IsBestDeal,
					"best_deal_started_at": p.BestDealStartedAt,
					"best_deal_expires_at": p.BestDealExpiresAt,
				})
				totalProductsUpdated++
			}

			productIndex++
		}
	}

	fmt.Printf("📊 Statistik Seeding: %d produk baru dibuat, %d produk diperbarui. Total: %d produk aktif!\n",
		totalProductsCreated, totalProductsUpdated, productIndex-1)
}
