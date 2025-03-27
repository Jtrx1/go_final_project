package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jtrx1/go_final_project/nextdate"
	"github.com/gin-gonic/gin"
)

// Структура для добавления задачи
type TaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title" binding:"required"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Структура для поиска задачи
type TaskResponse struct {
	ID      string `json:"id"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

func NextDateHandler(c *gin.Context) {
	nowStr := c.Query("now")
	dateStr := c.Query("date")
	repeat := c.Query("repeat")

	// Парсинг даты
	now, err := time.Parse(nextdate.TimeFormat, nowStr)
	if err != nil {
		c.String(http.StatusBadRequest, "Некорректный формат даты в параметре 'now'")
		return
	}

	// Вычисление следующей даты
	nextDate, err := nextdate.NextDate(now, dateStr, repeat)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	// Возвращаем только строку с датой
	c.String(http.StatusOK, nextDate)
}

func AddTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TaskRequest
		now := time.Now().UTC()

		// Парсинг JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса."})
			return
		}

		// Валидация обязательных полей
		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать заголовок задачи"})
			return
		}

		// Обработка даты
		var dateStr string
		if req.Date != "" {
			// Парсим входящую дату в UTC
			parsedDate, err := time.ParseInLocation(nextdate.TimeFormat, req.Date, time.UTC)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат даты"})
				return
			}
			if parsedDate.Before(now) {
				dateStr = now.Format(nextdate.TimeFormat)
			} else {
				dateStr = req.Date
			}
		} else {
			dateStr = now.Format(nextdate.TimeFormat)
		}
		// Вывод ошибки в случае некорректного правила повторения
		if req.Repeat != "" {
			_, err := nextdate.NextDate(now, dateStr, req.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}

		}
		result, err := db.Exec(
			"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
			dateStr,
			req.Title,
			req.Comment,
			req.Repeat,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения задачи"})
			return
		}
		id, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID задачи"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}
func GetTasks(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		search := strings.TrimSpace(c.Query("search"))
		tasks := make([]TaskResponse, 0)

		// Базовый запрос с сортировкой
		query := `
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            WHERE 1=1
        `
		args := []interface{}{}

		// Обработка поискового запроса
		if search != "" {
			// Попытка парсинга даты
			if t, err := time.Parse("02.01.2006", search); err == nil {
				query += " AND date = ?"
				args = append(args, t.Format("20060102"))
			} else {
				// Текстовый поиск с экранированием спецсимволов
				query += " AND (title LIKE ? ESCAPE '\\' OR comment LIKE ? ESCAPE '\\')"
				searchTerm := "%" + escapeLike(search) + "%"
				args = append(args, searchTerm, searchTerm)
			}
		}

		// Добавляем сортировку и лимит
		query += " ORDER BY date ASC, id ASC LIMIT 50"

		// Выполняем запрос с параметрами
		rows, err := db.Query(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка поиска задач"})
			return
		}
		defer rows.Close()

		// Обработка результатов
		for rows.Next() {
			var task TaskResponse
			var id int64

			if err := rows.Scan(
				&id,
				&task.Date,
				&task.Title,
				&task.Comment,
				&task.Repeat,
			); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки данных"})
				return
			}

			task.ID = strconv.FormatInt(id, 10)
			tasks = append(tasks, task)
		}

		if err = rows.Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения результатов"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}

}
func escapeLike(s string) string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(s, "\\", "\\\\"),
			"%", "\\%",
		),
		"_", "\\_",
	)
}
func GetTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан идентификатор"})
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор"})
			return
		}

		var task TaskResponse

		err = db.QueryRow(`
            SELECT id, date, title, comment, repeat 
            FROM scheduler 
            WHERE id = ?
        `, id).Scan(
			&task.ID,
			&task.Date,
			&task.Title,
			&task.Comment,
			&task.Repeat,
		)

		switch {
		case err == sql.ErrNoRows:
			c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
		case err != nil:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
		default:
			c.JSON(http.StatusOK, gin.H{
				"id":      task.ID,
				"date":    task.Date,
				"title":   task.Title,
				"comment": task.Comment,
				"repeat":  task.Repeat,
			})
		}
	}
}
func EditTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
