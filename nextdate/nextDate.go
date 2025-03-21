package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const TimeFormat = "20060102"

type Task struct {
	ID      int64  `json:"id,string"`
	Date    string `json:"date"`
	Title   string `json:"title"`
	Comment string `json:"comment"`
	Repeat  string `json:"repeat"`
}

// Представляем ответ API с возможными ошибками
type Response struct {
	ID    int64  `json:"id,omitempty,string"`
	Error string `json:"error,omitempty"`
}

// Представляем ответ API, содержащий список задач
type TaskListResponse struct {
	Tasks []Task `json:"tasks"`
}

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("Повторение не требуется")
	}

	// Разбираем начальную дату задачи
	parseDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", err
	}

	parts := strings.Split(repeat, "")
	if len(parts) > 2 {
		return "", fmt.Errorf("Некоррекный формат повторений")
	}

	switch parts[0] {
	case "d":
		return dayRepeat(now, parseDate, parts[1])
	case "y":
		return yearRepeat(now, parseDate), nil
	case "w":
		return weekRepeat(now, parseDate, parts[1])
	case "m":
	default:
		return "", fmt.Errorf("Некорректное правило повторения")
	}

}

func dayRepeat(now time.Time, parseDate time.Time, dayStr string) (string, error) {
	days, err := strconv.Atoi(dayStr)
	if err != nil {
		return "", fmt.Errorf("Ошибка формата дней")
	}
	if days >= 400 {
		return "", fmt.Errorf("Слишком большое количество дней")
	}

	for {
		parseDate = parseDate.AddDate(0, 0, days)
		if parseDate.After(now) {
			return parseDate.Format(TimeFormat), err
		}
	}
}
func yearRepeat(now time.Time, parseDate time.Time) (string, error) {
	parseDate = parseDate.AddDate(1, 0, 0)
	if parseDate.After(now) {
		return parseDate.Format(TimeFormat), nil
	}
	return "", fmt.Errorf("Ошибка при попытке добавить год")
}

func weekRepeat(now time.Time, parseDate time.Time, dayStr string) (string, error) {
	weekdays := make(map[time.Weekday]struct{})
	if dayStr == "" {
		return "", fmt.Errorf("Пустая строка дней недели")
	}

	for _, wd := range strings.Split(dayStr, ",") {
		wdNum, err := strconv.Atoi(wd)
		if err != nil || wdNum < 1 || wdNum > 7 {
			return "", fmt.Errorf("Неверный номер дня: %s (требуется 1-7)", wd)
		}
		weekdays[time.Weekday(wdNum%7)] = struct{}{}
	}

	currentDate := parseDate.AddDate(0, 0, 1)
	currentWeekday := currentDate.Weekday()

	minDays := 7
	for target := range weekdays {
		daysUntil := (int(target) - int(currentWeekday) + 7) % 7
		if daysUntil < minDays {
			minDays = daysUntil
		}
	}

	nextDate := currentDate.AddDate(0, 0, minDays)
	return nextDate.Format(TimeFormat), nil

}
func monthRepeat() {

}
