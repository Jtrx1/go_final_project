package scheduler

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type TaskResponse struct {
	ID      int64  `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

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

	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		file, err := os.Create(dbFile)
		if err != nil {
			log.Fatal("Ошибка создания файла: ", err)
		}
		log.Println("Файл успешно создан")
		err = os.Chmod(dbFile, 0777)
		if err != nil {
			log.Fatal("Ошибка установки прав на файл: ", err)
		}
		log.Println("Права на файл успешно установлены")

		file.Close()
	}
	// Открываем соединение с БД
	db, err := sql.Open("sqlite3", dbFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к БД: %w", err)
	}
	err = createTable(db)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания таблиц в БД: %w", err)
	}

	log.Println("База данных успешно инициализирована")
	return db, nil
}

func GetTaskDb(db *sql.DB, id int64) (*TaskResponse, int, error) {

	task := &TaskResponse{}

	err := db.QueryRow(`SELECT * FROM scheduler WHERE id = ?`, id).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat,
	)

	switch {
	case err == sql.ErrNoRows:
		return nil, http.StatusBadRequest, fmt.Errorf("Задача не найдена")
	case err != nil:
		return nil, http.StatusInternalServerError, fmt.Errorf("Ошибка базы данных")
	default:
		return task, http.StatusOK, nil
	}
}

func GetTasksDB(db *sql.DB, search string, isDate bool, limit int) ([]*TaskResponse, int, error) {
	tasks := make([]*TaskResponse, 0)

	var query string
	var args []any

	if isDate {
		query = `
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            WHERE date = ? 
            LIMIT ?
        `
		args = []any{search, limit}
	} else {
		query = `
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            WHERE title LIKE ? OR comment LIKE ? 
            ORDER BY date 
            LIMIT ?
        `
		search = "%" + search + "%"
		args = []any{search, search, limit}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, http.StatusInternalServerError, fmt.Errorf("Ошибка чтения данных: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		task := &TaskResponse{}
		err := rows.Scan(
			&task.ID,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat,
		)
		if err != nil {
			return nil, http.StatusInternalServerError, fmt.Errorf("Ошибка чтения данных: %w", err)
		}
		tasks = append(tasks, task)
	}
	return tasks, http.StatusInternalServerError, err
}

func DeleteTaskDB(db *sql.DB, id int64) (int, error) {
	result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Ошибка чтения данных: %w", err)
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return http.StatusNotFound, fmt.Errorf("Не удалено ни одной задачи")
	}
	return http.StatusOK, nil
}

func UpdateTaskDB(db *sql.DB, task TaskResponse) (int, error) {
	_, err := db.Exec(`
				UPDATE scheduler 
				SET date = ?, title = ?, comment = ?, repeat = ?
				WHERE id = ?`,
		task.Date,
		task.Title,
		task.Comment,
		task.Repeat,
		task.ID,
	)

	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("Ошибка обновления задачи: %w", err)
	}

	return http.StatusOK, nil
}
