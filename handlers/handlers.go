package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jtrx1/go_final_project/nextdate"
	"github.com/Jtrx1/go_final_project/scheduler"
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

		var isDate bool
		if search != "" {
			if t, err := time.Parse("02.01.2006", search); err == nil {
				search = t.Format("20060102")
				isDate = true
			}
		}
		tasks := make([]*scheduler.TaskResponse, 0)
		var code int
		var err error
		tasks, code, err = scheduler.GetTasksDB(db, search, isDate, 100)
		if err != nil {
			c.JSON(code, gin.H{"error": err})
		}
		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}

}

func GetTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		id, code, err := getTaskId(idStr)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
			return
		}
		task, statusCode, err := scheduler.GetTaskDb(db, id)
		if err != nil {
			c.JSON(statusCode, gin.H{"error": err.Error()})
			return
		} else {
			c.JSON(statusCode, gin.H{
				"id":      strconv.FormatInt(task.ID, 10),
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
		var req TaskResponse

		// Парсинг и валидация входных данных
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
			return
		}

		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для задачи обязателен заголовок"})
			return
		}
		// Обработка даты
		now := time.Now().UTC()
		dateStr := req.Date
		if req.Date != "" {
			parsedDate, err := time.Parse(nextdate.TimeFormat, req.Date)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат даты"})
				return
			}
			// Если дата в прошлом - использовать текущую
			if parsedDate.Before(now) {
				dateStr = now.Format(nextdate.TimeFormat)
			}
		} else {
			dateStr = now.Format(nextdate.TimeFormat)
		}
		// Обработка повторений
		if req.Repeat != "" {
			nextDate, err := nextdate.NextDate(now, dateStr, req.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			dateStr = nextDate
		}
		req.Date = dateStr
		var task scheduler.TaskResponse
		task.Comment = req.Comment
		task.Date = req.Date
		task.ID, _ = strconv.ParseInt(req.ID, 10, 64)
		task.Repeat = req.Repeat
		task.Title = req.Title

		code, err := scheduler.UpdateTaskDB(db, task)

		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func TaskDone(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		id, code, err := getTaskId(idStr)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
			return
		}
		task, code, err := scheduler.GetTaskDb(db, id)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
			return
		}
		if task.Repeat == "" {
			code, err = scheduler.DeleteTaskDB(db, id)
			if err != nil {
				c.JSON(code, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{})
			return
		} else {
			nextDate, err := nextdate.NextDate(time.Now(), task.Date, task.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			task.Date = nextDate
		}

		code, err = scheduler.UpdateTaskDB(db, *task)

		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
			return
		}

		c.JSON(code, gin.H{})
	}

}

func DeleteTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		id, code, err := getTaskId(idStr)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}
		code, err = scheduler.DeleteTaskDB(db, id)
		if err != nil {
			c.JSON(code, gin.H{"error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{})
	}
}

func getTaskId(idStr string) (int64, int, error) {
	if idStr == "" {
		return 0, http.StatusNotFound, fmt.Errorf("Не указан идентификатор задачи")
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return 0, http.StatusBadRequest, fmt.Errorf("Некорректный формат ID: %w", err)
	}

	return id, 200, nil

}
