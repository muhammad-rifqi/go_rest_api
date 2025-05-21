package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"github.com/gofiber/fiber/v2"
	_ "github.com/lib/pq" 
)

var db *sql.DB

func connectDB() (*sql.DB, error) {
	connStr := "user=postgres password=123 dbname=kneks sslmode=disable" 
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("gagal membuka koneksi database: %w", err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan ping ke database: %w", err)
	}

	fmt.Println("Berhasil terhubung ke database PostgreSQL!")
	return db, nil
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	var err error
	db, err = connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	app := fiber.New()
	app.Static("/", "./public")
	app.Get("/", home)
	app.Get("/users/:id", getUserByID)
	app.Get("/profile", profile)
	log.Fatal(app.Listen(":3000"))
}

func home(c *fiber.Ctx) error {
	rows, err := db.Query("SELECT id, username,password FROM users")
	if err != nil {
		return fmt.Errorf("gagal melakukan query: %w", err)
	}
	defer rows.Close()
	var users []struct {
		ID   int
		Username string
		Password string
	}
	for rows.Next() {
		var user struct {
			ID   int
			Username string
			Password string
		}
		if err := rows.Scan(&user.ID, &user.Username, &user.Password); err != nil {
			return fmt.Errorf("gagal memindai baris: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("terjadi kesalahan saat iterasi baris: %w", err)
	}
	return c.JSON(users)
	// return c.SendFile("./public/index.html")
}

func getUserByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("ID tidak valid")
	}

	var user User
	row := db.QueryRow("SELECT id, username, password FROM users WHERE id = $1", id)
	err = row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Pengguna dengan ID %d tidak ditemukan", id))
		}
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Gagal memindai pengguna: %v", err))
	}

	return c.JSON(user)
}

func profile(c *fiber.Ctx) error {
	return c.SendFile("./public/profile.html")
}