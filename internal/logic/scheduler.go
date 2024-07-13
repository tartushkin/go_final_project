package logic

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// CalculateNextDate вычисляет следующую дату для задачи в соответствии с указанным правилом.
func CalculateNextDate(now time.Time, date string, repeat string) (string, error) {
	if repeat == "" {
		return "", fmt.Errorf("пустое правило повторения")
	}

	// Парсим исходную дату
	taskDate, err := time.Parse("20060102", date)
	if err != nil {
		return "", fmt.Errorf("некорректная дата: %v", err)
	}

	// Обрабатываем правила повторения
	switch {
	case repeat == "y":
		// Ежегодное повторение
		for taskDate.Before(now) {
			taskDate = taskDate.AddDate(1, 0, 0)
		}
	case strings.HasPrefix(repeat, "d "):
		// Повторение через заданное количество дней
		days, err := strconv.Atoi(strings.TrimPrefix(repeat, "d "))
		if err != nil || days < 1 || days > 400 {
			return "", fmt.Errorf("некорректный формат повторения: %v", repeat)
		}
		for taskDate.Before(now) {
			taskDate = taskDate.AddDate(0, 0, days)
		}
	case strings.HasPrefix(repeat, "w "):
		// Повторение по дням недели
		daysOfWeek := strings.Split(strings.TrimPrefix(repeat, "w "), ",")
		for {
			for _, day := range daysOfWeek {
				dayInt, err := strconv.Atoi(day)
				if err != nil || dayInt < 1 || dayInt > 7 {
					return "", fmt.Errorf("некорректный день недели: %v", day)
				}
				weekday := time.Weekday(dayInt - 1)
				if taskDate.Weekday() == weekday && taskDate.After(now) {
					return taskDate.Format("20060102"), nil
				}
			}
			taskDate = taskDate.AddDate(0, 0, 1)
		}
	case strings.HasPrefix(repeat, "m "):
		// Повторение по дням месяца
		parts := strings.Split(strings.TrimPrefix(repeat, "m "), " ")
		if len(parts) == 0 || len(parts) > 2 {
			return "", fmt.Errorf("некорректный формат повторения: %v", repeat)
		}

		daysOfMonth := strings.Split(parts[0], ",")
		var months []string
		if len(parts) == 2 {
			months = strings.Split(parts[1], ",")
		}

		for {
			for _, day := range daysOfMonth {
				dayInt, err := strconv.Atoi(day)
				if err != nil || dayInt < -2 || dayInt > 31 {
					return "", fmt.Errorf("некорректный день месяца: %v", day)
				}

				var newDate time.Time
				if dayInt > 0 {
					newDate = time.Date(taskDate.Year(), taskDate.Month(), dayInt, 0, 0, 0, 0, taskDate.Location())
				} else {
					newDate = time.Date(taskDate.Year(), taskDate.Month()+1, dayInt, 0, 0, 0, 0, taskDate.Location()).AddDate(0, 0, -1)
				}

				if len(months) > 0 {
					for _, month := range months {
						monthInt, err := strconv.Atoi(month)
						if err != nil || monthInt < 1 || monthInt > 12 {
							return "", fmt.Errorf("некорректный месяц: %v", month)
						}
						if newDate.Month() == time.Month(monthInt) && newDate.After(now) {
							return newDate.Format("20060102"), nil
						}
					}
				} else if newDate.After(now) {
					return newDate.Format("20060102"), nil
				}
			}
			taskDate = taskDate.AddDate(0, 1, 0)
		}
	default:
		return "", fmt.Errorf("неподдерживаемый формат: %v", repeat)
	}

	return taskDate.Format("20060102"), nil
}
