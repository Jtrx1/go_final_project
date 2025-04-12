package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const TimeFormat = "20060102"

func NextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("повторение не требуется")
	}

	// Разбираем начальную дату задачи
	parseDate, err := time.Parse(TimeFormat, date)
	if err != nil {
		return "", err
	}

	parts := strings.Split(repeat, " ")

	switch parts[0] {
	case "d":
		if len(parts) < 2 { // Если нет второй части
			return "", fmt.Errorf("некорректный формат правила повторения")
		}
		return dayRepeat(now, parseDate, parts[1])
	case "y":
		return yearRepeat(now, parseDate)
	default:
		return "", fmt.Errorf("некорректное правило повторения")
	}

}

func dayRepeat(now time.Time, parseDate time.Time, dayStr string) (string, error) {
	days, err := strconv.Atoi(dayStr)
	if err != nil || days <= 0 {
		return "", fmt.Errorf("ошибка формата дней")
	}
	if days >= 400 {
		return "", fmt.Errorf("слишком большое количество дней")
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
