package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func ConnectDB() error {
	connStr := "user=postgres password=123 dbname=kneks sslmode=disable"
	var err error
	DB, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	err = DB.Ping()
	if err != nil {
		return fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	fmt.Println("Berhasil terhubung ke database PostgreSQL!")
	return nil
}
