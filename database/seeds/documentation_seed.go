package seeds

import (
	"log"
	"nexfi-backend/models"
	"time"

	"gorm.io/gorm"
)

// SeedDocumentation seeds sample documentation data
func SeedDocumentation(db *gorm.DB) error {
	// Check if categories already exist
	var count int64
	db.Model(&models.DocCategory{}).Count(&count)
	if count > 0 {
		log.Println("📚 Documentation data already exists, skipping seed...")
		return nil
	}

	log.Println("📚 Seeding documentation data...")

	// Create categories
	categories := []models.DocCategory{
		{
			Name:        "Getting Started",
			Slug:        "getting-started",
			Description: "Pelajari dasar-dasar NexFlux dan mulai perjalananmu dalam dunia elektronika",
			Icon:        "rocket",
			Color:       "from-brand-500 to-neon-500",
			Order:       1,
			IsActive:    true,
		},
		{
			Name:        "Hardware Guides",
			Slug:        "hardware",
			Description: "Panduan lengkap untuk berbagai komponen hardware dan mikrokontroler",
			Icon:        "microchip",
			Color:       "from-blue-500 to-cyan-500",
			Order:       2,
			IsActive:    true,
		},
		{
			Name:        "Programming",
			Slug:        "programming",
			Description: "Tutorial pemrograman untuk Arduino, ESP32, dan platform IoT lainnya",
			Icon:        "code",
			Color:       "from-purple-500 to-pink-500",
			Order:       3,
			IsActive:    true,
		},
		{
			Name:        "Projects",
			Slug:        "projects",
			Description: "Koleksi project menarik yang bisa kamu buat dengan NexFlux",
			Icon:        "lightbulb",
			Color:       "from-yellow-500 to-orange-500",
			Order:       4,
			IsActive:    true,
		},
		{
			Name:        "Troubleshooting",
			Slug:        "troubleshooting",
			Description: "Solusi untuk masalah umum dan tips debugging",
			Icon:        "bug",
			Color:       "from-red-500 to-rose-500",
			Order:       5,
			IsActive:    true,
		},
	}

	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			return err
		}
	}

	now := time.Now()

	// Create articles
	articles := []models.DocArticle{
		// Getting Started articles
		{
			CategoryID:      categories[0].ID,
			Title:           "Cara Memulai dengan Arduino Uno",
			Slug:            "memulai-arduino-uno",
			Excerpt:         "Panduan lengkap untuk pemula yang ingin memulai perjalanan dengan Arduino Uno. Pelajari dasar-dasar, setup, dan program pertama Anda.",
			Content:         getArduinoStarterContent(),
			ReadTimeMinutes: 10,
			Difficulty:      models.DocDifficultyBeginner,
			Tags:            []string{"arduino", "beginner", "tutorial", "uno"},
			Views:           15420,
			IsFeatured:      true,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		{
			CategoryID:      categories[0].ID,
			Title:           "Mengenal Interface NexFlux",
			Slug:            "mengenal-interface-nexflux",
			Excerpt:         "Tour lengkap untuk memahami semua fitur dan interface di platform NexFlux.",
			Content:         "# Mengenal Interface NexFlux\n\nSelamat datang di NexFlux! Mari kita pelajari interface platform ini.\n\n## Dashboard\n\nDashboard adalah halaman utama tempat kamu melihat progress...\n\n## Studio\n\nStudio adalah tempat kamu membuat dan mengedit project elektronika...",
			ReadTimeMinutes: 8,
			Difficulty:      models.DocDifficultyBeginner,
			Tags:            []string{"nexflux", "interface", "tutorial"},
			Views:           8950,
			IsFeatured:      true,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		// Hardware Guides
		{
			CategoryID:      categories[1].ID,
			Title:           "Panduan Lengkap LED dan Resistor",
			Slug:            "panduan-led-resistor",
			Excerpt:         "Pelajari cara menggunakan LED dengan resistor yang tepat untuk menghindari kerusakan komponen.",
			Content:         "# Panduan LED dan Resistor\n\nLED (Light Emitting Diode) adalah komponen dasar...\n\n## Menghitung Resistor\n\nGunakan rumus: R = (Vs - Vf) / If\n\n- Vs = Tegangan sumber\n- Vf = Forward voltage LED (~2V untuk LED biasa)\n- If = Forward current (~20mA)",
			ReadTimeMinutes: 7,
			Difficulty:      models.DocDifficultyBeginner,
			Tags:            []string{"led", "resistor", "basic"},
			Views:           12300,
			IsFeatured:      false,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		{
			CategoryID:      categories[1].ID,
			Title:           "ESP32: Panduan WiFi dan Bluetooth",
			Slug:            "esp32-wifi-bluetooth",
			Excerpt:         "Manfaatkan kemampuan wireless ESP32 untuk project IoT yang powerful.",
			Content:         "# ESP32: WiFi dan Bluetooth\n\nESP32 adalah mikrokontroler yang powerful dengan built-in WiFi dan Bluetooth...\n\n## Koneksi WiFi\n\n```cpp\n#include <WiFi.h>\n\nconst char* ssid = \"your-ssid\";\nconst char* password = \"your-password\";\n\nvoid setup() {\n  WiFi.begin(ssid, password);\n}\n```",
			ReadTimeMinutes: 15,
			Difficulty:      models.DocDifficultyIntermediate,
			Tags:            []string{"esp32", "wifi", "bluetooth", "iot"},
			Views:           9800,
			IsFeatured:      true,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		// Programming
		{
			CategoryID:      categories[2].ID,
			Title:           "Dasar Pemrograman Arduino",
			Slug:            "dasar-pemrograman-arduino",
			Excerpt:         "Pelajari struktur program Arduino: setup(), loop(), dan fungsi-fungsi dasar lainnya.",
			Content:         "# Dasar Pemrograman Arduino\n\n## Struktur Program\n\n```cpp\nvoid setup() {\n  // Dijalankan sekali saat pertama kali\n}\n\nvoid loop() {\n  // Dijalankan terus-menerus\n}\n```\n\n## Digital I/O\n\n- `pinMode(pin, mode)` - Set mode pin\n- `digitalWrite(pin, value)` - Tulis nilai digital\n- `digitalRead(pin)` - Baca nilai digital",
			ReadTimeMinutes: 12,
			Difficulty:      models.DocDifficultyBeginner,
			Tags:            []string{"arduino", "programming", "cpp"},
			Views:           18500,
			IsFeatured:      true,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		{
			CategoryID:      categories[2].ID,
			Title:           "PWM dan Analog Output",
			Slug:            "pwm-analog-output",
			Excerpt:         "Kontrol kecerahan LED dan kecepatan motor dengan PWM (Pulse Width Modulation).",
			Content:         "# PWM dan Analog Output\n\n## Apa itu PWM?\n\nPWM adalah teknik untuk mensimulasikan output analog...\n\n## Contoh Penggunaan\n\n```cpp\nint ledPin = 9;\nint brightness = 0;\n\nvoid loop() {\n  analogWrite(ledPin, brightness);\n  brightness = (brightness + 5) % 256;\n  delay(30);\n}\n```",
			ReadTimeMinutes: 10,
			Difficulty:      models.DocDifficultyIntermediate,
			Tags:            []string{"pwm", "analog", "motor", "led"},
			Views:           7200,
			IsFeatured:      false,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		// Projects
		{
			CategoryID:      categories[3].ID,
			Title:           "Project: Smart Home Automation",
			Slug:            "project-smart-home",
			Excerpt:         "Buat sistem home automation sederhana dengan ESP32 dan relay module.",
			Content:         "# Smart Home Automation\n\nDalam tutorial ini kita akan membuat sistem smart home...\n\n## Komponen yang Dibutuhkan\n\n- ESP32\n- Relay Module 4 channel\n- Kabel jumper\n- Power supply 5V\n\n## Wiring Diagram\n\n[Diagram akan ditampilkan di sini]",
			ReadTimeMinutes: 25,
			Difficulty:      models.DocDifficultyAdvanced,
			Tags:            []string{"project", "smart-home", "esp32", "iot"},
			Views:           5600,
			IsFeatured:      true,
			IsPublished:     true,
			PublishedAt:     &now,
		},
		// Troubleshooting
		{
			CategoryID:      categories[4].ID,
			Title:           "Error Upload Arduino: Solusi Lengkap",
			Slug:            "error-upload-arduino",
			Excerpt:         "Kumpulan solusi untuk masalah upload program ke Arduino yang sering ditemui.",
			Content:         "# Error Upload Arduino\n\n## Error: avrdude stk500_getsync\n\nPenyebab:\n- Port tidak tepat\n- Driver belum terinstall\n\nSolusi:\n1. Pastikan pilih port yang benar di Tools > Port\n2. Install driver CH340 jika menggunakan clone Arduino\n\n## Error: Programmer not responding\n\nCoba:\n1. Tekan tombol reset saat upload\n2. Ganti kabel USB",
			ReadTimeMinutes: 8,
			Difficulty:      models.DocDifficultyBeginner,
			Tags:            []string{"troubleshooting", "error", "arduino"},
			Views:           21000,
			IsFeatured:      false,
			IsPublished:     true,
			PublishedAt:     &now,
		},
	}

	for i := range articles {
		if err := db.Create(&articles[i]).Error; err != nil {
			return err
		}
	}

	// Create videos
	videos := []models.DocVideo{
		{
			CategoryID:      categories[0].ID,
			Title:           "Arduino untuk Pemula - Tutorial Lengkap 30 Menit",
			Description:     "Video tutorial komprehensif untuk pemula yang ingin belajar Arduino dari nol.",
			VideoURL:        "https://www.youtube.com/watch?v=nL34zDTPkcs",
			ThumbnailURL:    "https://img.youtube.com/vi/nL34zDTPkcs/maxresdefault.jpg",
			DurationSeconds: 1935,
			Difficulty:      models.DocDifficultyBeginner,
			Views:           8500,
			IsFeatured:      true,
			IsPublished:     true,
		},
		{
			CategoryID:      categories[1].ID,
			Title:           "Cara Menggunakan Multimeter untuk Pemula",
			Description:     "Pelajari cara mengukur tegangan, arus, dan resistansi dengan multimeter.",
			VideoURL:        "https://www.youtube.com/watch?v=SLkPtmnglOI",
			ThumbnailURL:    "https://img.youtube.com/vi/SLkPtmnglOI/maxresdefault.jpg",
			DurationSeconds: 1200,
			Difficulty:      models.DocDifficultyBeginner,
			Views:           5200,
			IsFeatured:      false,
			IsPublished:     true,
		},
		{
			CategoryID:      categories[2].ID,
			Title:           "Membuat LED Blink dengan Kode",
			Description:     "Tutorial coding pertama: membuat LED berkedip dengan Arduino.",
			VideoURL:        "https://www.youtube.com/watch?v=fJWR7dBuc18",
			ThumbnailURL:    "https://img.youtube.com/vi/fJWR7dBuc18/maxresdefault.jpg",
			DurationSeconds: 600,
			Difficulty:      models.DocDifficultyBeginner,
			Views:           12000,
			IsFeatured:      true,
			IsPublished:     true,
		},
	}

	for i := range videos {
		if err := db.Create(&videos[i]).Error; err != nil {
			return err
		}
	}

	log.Printf("✅ Seeded %d categories, %d articles, %d videos", len(categories), len(articles), len(videos))
	return nil
}

func getArduinoStarterContent() string {
	return `# Cara Memulai dengan Arduino Uno

Selamat datang di dunia Arduino! Panduan ini akan membantu Anda memulai perjalanan dengan Arduino Uno.

## Apa itu Arduino Uno?

Arduino Uno adalah board mikrokontroler berbasis ATmega328P yang populer untuk pembelajaran dan prototyping elektronika.

## Persiapan

Sebelum memulai, pastikan Anda memiliki:
- Arduino Uno board
- Kabel USB Type-B
- Komputer dengan Windows/Mac/Linux
- Arduino IDE (download dari arduino.cc)

## Instalasi Arduino IDE

1. Kunjungi [arduino.cc/software](https://www.arduino.cc/en/software)
2. Download versi sesuai OS Anda
3. Install dan jalankan Arduino IDE

## Program Pertama: Blink LED

Mari buat LED built-in berkedip!

### Kode

` + "```cpp" + `
void setup() {
  pinMode(LED_BUILTIN, OUTPUT);
}

void loop() {
  digitalWrite(LED_BUILTIN, HIGH);
  delay(1000);
  digitalWrite(LED_BUILTIN, LOW);
  delay(1000);
}
` + "```" + `

### Cara Upload

1. Hubungkan Arduino ke komputer via USB
2. Pilih board: Tools > Board > Arduino Uno
3. Pilih port: Tools > Port > (pilih port Arduino)
4. Klik tombol Upload (panah kanan)

## Selamat!

LED built-in pada Arduino Anda seharusnya sudah berkedip setiap 1 detik.

## Langkah Selanjutnya

- Pelajari [Dasar Pemrograman Arduino](/docs/programming/dasar-pemrograman-arduino)
- Coba [Panduan LED dan Resistor](/docs/hardware/panduan-led-resistor)
`
}
