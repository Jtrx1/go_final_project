package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const TimeFormat = "20060102"

// Представляем ответ API с возможными ошибками
type Response struct {
	ID    int64  `json:"id,omitempty,string"`
	Error string `json:"error,omitempty"`
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

	parts := strings.Split(repeat, " ")

	switch parts[0] {
	case "d":
		return dayRepeat(now, parseDate, parts[1])
	case "y":
		return yearRepeat(now, parseDate)
	default:
		return "", fmt.Errorf("Некорректное правило повторения")
	}

}

func dayRepeat(now time.Time, parseDate time.Time, dayStr string) (string, error) {
	days, err := strconv.Atoi(dayStr)
	if err != nil || days <= 0 {
		return "", fmt.Errorf("Ошибка формата дней")
	}
	if days >= 400 {
		return "", fmt.Errorf("Слишком большое количество дней")
	}

	for {
		parseDate = parseDate.AddDate(0, 0, days)
		if parseDate.After(now) {
			return parseDate.Format(TimeFormat), nil
		}
	}
}
func yearRepeat(now time.Time, parseDate time.Time) (string, error) {
	// Добавляем годы, пока не превысим текущую дату
	for {
		parseDate = parseDate.AddDate(1, 0, 0)
		if parseDate.After(now) {
			return parseDate.Format(TimeFormat), nil
		}
	}
}
