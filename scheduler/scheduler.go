package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

func createTable(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS scheduler (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            date CHAR(8) NOT NULL,
            title TEXT NOT NULL,
            comment TEXT,
            repeat VARCHAR(128)
        );`,
		`CREATE INDEX IF NOT EXISTS scheduler_date ON scheduler (date);`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("ошибка выполнения запроса %q: %w", query, err)
		}
	}
	return nil
}

func InitDB(dbFile string) (*sql.DB, error) {
	// Создаём каталог для БД, если он не существует
	dir := filepath.Dir(dbFile)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("не удалось создать каталог %q: %w", dir, err)
	}

	// Открываем соединение с БД
	db, err := sql.Open("sqlite3", dbFile)
	createTable(db)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}
	log.Println("База данных успешно инициализирована")
	return db, nil
}
