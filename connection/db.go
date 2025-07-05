package connection

import (
	"al/models"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("❗ Gagal mendapatkan data file .env")
	}

	// Baca lokasi file SQLite dari .env, default: "./database.sqlite"
	dbFile := os.Getenv("DB_FILE")
	if dbFile == "" {
		dbFile = "./database.sqlite"
	}

	fmt.Println("Menggunakan file SQLite database:", dbFile)

	// Buka koneksi dengan SQLite
	db, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		log.Fatal("💥 Koneksi Database SQLite Gagal, error : ", err)
	}

	DB = db

	// Print stats koneksi (optional)
	sqlDB, err := db.DB()
	if err == nil {
		fmt.Println("🎯 Real connection:", sqlDB.Stats())
	}

	// SQLite tidak punya konsep "current_database", jadi bisa skip ini
	// fmt.Println("✅ Connected to DB:", dbName)
}

// CleanUp menghapus data pada tabel dan pivot
func CleanUp(db *gorm.DB) error {
	// Contoh hapus isi tabel pivot many2many
	if err := db.Exec("DELETE FROM role_permissions").Error; err != nil {
		return err
	}

	// Hapus data di tabel utama
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Role{}).Error; err != nil {
		return err
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Permission{}).Error; err != nil {
		return err
	}
	if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.User{}).Error; err != nil {
		return err
	}

	return nil
}
