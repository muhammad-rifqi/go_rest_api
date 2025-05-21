package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"github.com/gofiber/fiber/v2" 
	"github.com/muhammad-rifqi/go_rest_api/models"
	"github.com/muhammad-rifqi/go_rest_api/database"
)

func main() {
	err := database.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer database.DB.Close()

	app := fiber.New()
	app.Static("/", "./public")
	app.Get("/", welcome)
	app.Get("/users", home)
	app.Get("/users/:id", getUserByID)
	app.Post("/users", createUser)
	app.Put("/users/:id", updateUser)
	app.Delete("/users/:id", deleteUser)
	app.Post("/login", loginUser)
	app.Get("/profile", profile)
	log.Fatal(app.Listen(":3000"))
}

func home(c *fiber.Ctx) error {
	rows, err := database.DB.Query("SELECT id, username,password FROM users")
	if err != nil {
		return fmt.Errorf("gagal melakukan query: %w", err)
	}
	defer rows.Close()
	var users []models.User
	for rows.Next() {
		var user models.User
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

	var user models.User
	row := database.DB.QueryRow("SELECT id, username, password FROM users WHERE id = $1", id)
	err = row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).SendString(fmt.Sprintf("Pengguna dengan ID %d tidak ditemukan", id))
		}
		return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Gagal memindai pengguna: %v", err))
	}

	return c.JSON(user)
}

func createUser(c *fiber.Ctx) error {
	var user models.User
	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal mem-parsing body",
		})
	}

	if user.Username == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username dan Password tidak boleh kosong",
		})
	}

	query := "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id"
	err := database.DB.QueryRow(query, user.Username, user.Password).Scan(&user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Gagal menyimpan user: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(user)
}

func deleteUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	result, err := database.DB.Exec("DELETE FROM users WHERE id = $1", id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Gagal menghapus user: %v", err),
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membaca hasil penghapusan",
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("User dengan ID %d tidak ditemukan", id),
		})
	}

	return c.JSON(fiber.Map{
		"message": fmt.Sprintf("User dengan ID %d berhasil dihapus", id),
	})
}

func updateUser(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ID tidak valid",
		})
	}

	var user models.User

	if err := c.BodyParser(&user); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal mem-parsing body",
		})
	}

	if user.Username == "" || user.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username dan Password tidak boleh kosong",
		})
	}

	query := "UPDATE users SET username = $1, password = $2 WHERE id = $3"
	result, err := database.DB.Exec(query, user.Username, user.Password, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Gagal mengupdate user: %v", err),
		})
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal membaca hasil update",
		})
	}

	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": fmt.Sprintf("User dengan ID %d tidak ditemukan", id),
		})
	}

	user.ID = id
	return c.JSON(fiber.Map{
		"message": "User berhasil diupdate",
		"user":    user,
	})
}

func loginUser(c *fiber.Ctx) error {
	var input struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&input); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Gagal mem-parsing body",
		})
	}

	if input.Username == "" || input.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Username dan password wajib diisi",
		})
	}

	var user models.User
	query := "SELECT id, username, password FROM users WHERE username = $1"
	err := database.DB.QueryRow(query, input.Username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Username tidak ditemukan",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Gagal mengambil data pengguna",
		})
	}

	if input.Password != user.Password {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Password salah",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Login berhasil",
		"user":    user,
	})
}


func profile(c *fiber.Ctx) error {
	return c.SendFile("./public/profile.html")
}

func welcome(c *fiber.Ctx) error {
	return c.SendFile("./public/welcome.html")
}