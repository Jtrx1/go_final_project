package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Jtrx1/go_final_project/nextdate"
	"github.com/Jtrx1/go_final_project/scheduler"
	"github.com/gin-gonic/gin"
)

// Структура для поиска задачи
type Task struct {
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
	c.String(http.StatusOK, nextDate)
}

func AddTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req scheduler.TaskResponse
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
		req.Date = dateStr
		id, err := scheduler.InsertTaskDB(db, req)
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
				search = t.Format(nextdate.TimeFormat)
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
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан идентификатор"})
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор"})
			return
		}

		var task scheduler.TaskResponse
		var code int
		task, code, err = scheduler.GetTaskDb(db, id)

		switch {
		case err != nil:
			c.JSON(code, gin.H{"error": err})
		default:
			c.JSON(http.StatusOK, task)
		}
	}
}
func EditTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req scheduler.TaskResponse

		// Парсинг и валидация входных данных
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
			return
		}

		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для задачи обязателен щаголовок"})
			return
		}

		// Проверка существования задачи
		exists, err := scheduler.TaskExists(db, req.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			return
		}
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
			return
		}

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
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное правило повторения"})
				return
			}
			dateStr = nextDate
		}

		req.Date = dateStr
		_, err = scheduler.UpdateTaskDB(db, req)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}

func TaskDone(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем и проверяем ID задачи
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан идентификатор задачи"})
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор задачи"})
			return
		}

		task, _, err := scheduler.GetTaskDb(db, id)

		if err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			}
			return
		}
		now := time.Now().UTC()

		// Обработка повторяющейся задачи
		if task.Repeat != "" {
			nextDate, err := nextdate.NextDate(now, task.Date, task.Repeat)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка вычисления даты: " + err.Error()})
				return
			}

			task.Date = nextDate
			_, err = scheduler.UpdateTaskDB(db, task)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
				return
			}
		} else {
			// Удаление одноразовой задачи
			_, err = scheduler.DeleteTaskDB(db, task.ID)

			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления задачи"})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{})
	}
}

func DeleteTask(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		idStr := c.Query("id")
		if idStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан идентификатор задачи"})
			return
		}

		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор задачи"})
			return
		}

		_, err = scheduler.DeleteTaskDB(db, id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}
