package main

import (
	"database/sql"
	"log"

	"github.com/gofiber/fiber/v2"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spacetab-io/my-bank-service/internal/account"
)

const (
	// Better to move this into config
	SqliteDBPath = "./bank.db"
	AppPort      = ":8080"
)

func main() {
	// Database Init
	db, err := sql.Open("sqlite3", SqliteDBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Migrations
	err = account.LoadMigrations(db)
	if err != nil {
		log.Fatal(err)
	}

	// Services
	as := account.NewService(db)

	// Routes
	app := fiber.New()
	account.RegisterHandlers(app, as)

	log.Fatal(app.Listen(AppPort))
}
