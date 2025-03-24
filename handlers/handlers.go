package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
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
		rows, err := db.Query(`
		SELECT id, date, title, comment, repeat 
		FROM scheduler 
		ORDER BY date ASC
	`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения задач"})
			return
		}
		defer rows.Close()

		var tasks []TaskResponse

		for rows.Next() {
			var task TaskResponse
			var id int64

			err := rows.Scan(
				&id,
				&task.Date,
				&task.Title,
				&task.Comment,
				&task.Repeat,
			)
			if err != nil {
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
