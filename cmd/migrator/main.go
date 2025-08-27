package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	"log"
	"os"
)

func main() {
	var migrationsPath, migrationsTable string

	flag.StringVar(&migrationsPath, "migrations-path", "./migrations", "Path to migrations folder")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "Table name")
	flag.Parse()

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

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
