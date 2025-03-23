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
		now := time.Now().UTC()

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
			return
		}

		// Валидация заголовка
		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать заголовок"})
			return
		}

		// Обработка даты
		var dateStr string
		if req.Date != "" {
			// Парсинг входящей даты
			parsedDate, err := time.Parse(nextdate.TimeFormat, req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат даты"})
				return
			}
			dateStr = parsedDate.Format(nextdate.TimeFormat)
		} else {
			// Дата по умолчанию - сегодня
			dateStr = now.Format(nextdate.TimeFormat)
		}

		// Корректировка даты для повторяющихся задач
		if req.Repeat != "" {
			// Рассчитываем следующую валидную дату
			nextDate, err := nextdate.NextDate(now, dateStr, req.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное правило повторения: " + err.Error()})
				return
			}
			dateStr = nextDate
		} else {
			// Проверка даты для не повторяющихся задач
			if dateStr < now.Format(nextdate.TimeFormat) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Дата не может быть меньше текущей"})
				return
			}
		}

		// Вставка в БД с корректной датой
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

		// Получение ID созданной записи
		id, err := result.LastInsertId()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID задачи"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"id": id})
	}
}
