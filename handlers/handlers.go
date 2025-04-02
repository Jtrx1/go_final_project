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
            SELECT * 
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
            SELECT * 
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
		var req TaskResponse

		// Парсинг и валидация входных данных
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат запроса"})
			return
		}

		if req.Title == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Для задачи обязателен щаголовок"})
			return
		}
		// Преобразование ID в число
		id, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный идентификатор"})
			return
		}

		// Проверка существования задачи
		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM scheduler WHERE id = ?)", id).Scan(&exists)
		if err != nil || !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
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
				c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректное правило повторения"})
				return
			}
			dateStr = nextDate
		}

		// Обновление в БД
		_, err = db.Exec(`
				UPDATE scheduler 
				SET date = ?, title = ?, comment = ?, repeat = ?
				WHERE id = ?`,
			dateStr,
			req.Title,
			req.Comment,
			req.Repeat,
			id,
		)

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

		// Начинаем транзакцию
		tx, err := db.Begin()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка начала транзакции"})
			return
		}
		defer tx.Rollback()

		// Получаем текущие данные задачи
		var currentDate, repeatRule string
		err = tx.QueryRow(`
            SELECT date, repeat 
            FROM scheduler 
            WHERE id = ?`,
			id).Scan(&currentDate, &repeatRule)

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
		if repeatRule != "" {
			nextDate, err := nextdate.NextDate(now, currentDate, repeatRule)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Ошибка вычисления даты: " + err.Error()})
				return
			}

			_, err = tx.Exec(`
                UPDATE scheduler 
                SET date = ?
                WHERE id = ?`,
				nextDate,
				id,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обновления задачи"})
				return
			}
		} else {
			// Удаление одноразовой задачи
			_, err = tx.Exec(`
                DELETE FROM scheduler 
                WHERE id = ?`,
				id,
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления задачи"})
				return
			}
		}

		// Фиксация транзакции
		if err := tx.Commit(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сохранения изменений"})
			return
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

		result, err := db.Exec("DELETE FROM scheduler WHERE id = ?", id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка базы данных"})
			return
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			c.JSON(http.StatusNotFound, gin.H{"error": "Задача не найдена"})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	}
}
