package config

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB() {
	var err error

	dbPath := getDBPath()

	log.Println("Using DB path:", dbPath)

	DB, err = sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatalf("failed to open DB: %v", err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatalf("failed to connect DB: %v", err)
	}
}

func getDBPath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalf("failed to get user config dir: %v", err)
	}

	appDir := filepath.Join(dir, "Fire")

	// buat folder kalau belum ada
	if err := os.MkdirAll(appDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create app dir: %v", err)
	}

	return filepath.Join(appDir, "fire.db")
}
