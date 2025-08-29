package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"github.com/joho/godotenv"
	"log"
	"os"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var migrationsPath, migrationsTable string

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "Path to migrations folder")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Table name")
	flag.Parse()

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: no .env file found: %v", err)
	}

	dbUser := os.Getenv("POSTGRES_USER")
	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	dbHost := os.Getenv("POSTGRES_HOST")
	dbPort := os.Getenv("POSTGRES_PORT")
	dbName := os.Getenv("POSTGRES_DB")

	fmt.Println(dbUser, dbPassword, dbHost, dbPort, dbName, migrationsPath, migrationsTable)

	// Формируем строку подключения
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&x-migrations-table=%s",
		dbUser, dbPassword, dbHost, dbPort, dbName, migrationsTable)
	log.Printf("Connecting to database with URL: %s", dbURL)

	// Выполняем миграции
	m, err := migrate.New("file://"+migrationsPath, dbURL)
	if err != nil {
		log.Println("migrate.New")
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			fmt.Println("no migrations to apply")
			return
		}
		panic(err)
	}

	fmt.Println("migrations applied successfully")
}
