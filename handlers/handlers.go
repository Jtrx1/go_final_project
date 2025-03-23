package handlers

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/Jtrx1/go_final_project/nextdate"
	"github.com/gin-gonic/gin"
)

type TaskRequest struct {
	Date    string `json:"date"`
	Title   string `json:"title" binding:"required"`
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

		// Парсинг JSON
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
			return
		}

		// Валидация обязательных полей
		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать заголовок"})
			return
		}

		// Обработка даты
		var date string
		if req.Date != "" {
			_, err := time.Parse(nextdate.TimeFormat, req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат даты"})
				return
			}
			date = req.Date
		} else {
			date = time.Now().Format(nextdate.TimeFormat)
		}

		// Валидация правила повтора
		if req.Repeat != "" {
			_, err := nextdate.NextDate(time.Now(), date, req.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное правило повторения: " + err.Error()})
				return
			}
		}

		// Вставка в БД
		result, err := db.Exec(
			"INSERT INTO scheduler (date, title, comment, repeat) VALUES (?, ?, ?, ?)",
			date,
			req.Title,
			req.Comment,
			req.Repeat,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения задачи"})
			return
		}

		// Получение ID созданной записи
		id, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID задачи"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}
