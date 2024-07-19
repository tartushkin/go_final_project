package http

import (
	"go_final_project/internal/app"
	"go_final_project/internal/date"
	"net/http"

	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

const (
	t = app.FormatDate
)

func validateID(id string) error {
	if id == "" {
		logrus.Debug("ID пустой")
		return echo.NewHTTPError(http.StatusBadRequest, "Не указан идентификатор")
	}
	logrus.Debugf("ID валиден: %s", id)
	return nil
}

func validateTitle(title string) error {
	if title == "" {
		logrus.Debug("Заголовок пустой")
		return echo.NewHTTPError(http.StatusBadRequest, "Не указан заголовок задачи")
	}
	logrus.Debugf("Заголовок валиден: %s", title)
	return nil
}

func validateDate(dateStr string) (time.Time, error) {
	logrus.Debugf("Проверка формата даты: %s", dateStr)
	if dateStr == "" {
		logrus.Debug("Дата не указана, используем сегодняшнюю дату")
		return time.Now().UTC(), nil
	}
	date, err := time.Parse(t, dateStr)
	if err != nil {
		logrus.Errorf("Некорректная дата: %v", err)
		return time.Time{}, fmt.Errorf("Некорректная дата: %v", err)
	}
	logrus.Debugf("Дата успешно разобрана: %v", date.UTC())
	return date.UTC(), nil
}

func validateRepeatDate(taskDate time.Time, repeat string) (string, error) {
	logrus.Debugf("Проверка даты и правила повторения: дата=%v, повторение=%s", taskDate, repeat)
	// Получаем текущую дату без времени

	now := time.Now().UTC().Truncate(24 * time.Hour)
	taskDate = taskDate.UTC().Truncate(24 * time.Hour)

	if taskDate.Before(now) {
		if repeat == "" {
			logrus.Debugf("Дата прошедшая и нет правила повторения, используем сегодняшнюю дату, дата= %v", now)
			return now.Format(t), nil
		}

		logrus.Debugf("Дата прошедшая, вычисление следующей даты по правилу повторения: %s", repeat)
		nextDate, err := date.CalculateNextDate(now, taskDate.Format(t), repeat)
		if err != nil {
			logrus.Errorf("Ошибка расчета следующей даты: %v", err)
			return "", fmt.Errorf("Ошибка расчета следующей даты: %v", err)
		}
		logrus.Debugf("Следующая дата успешно рассчитана: %s", nextDate)
		return nextDate, nil
	}

	logrus.Debugf("Дата действительная и не требует повторения: %v", taskDate)
	return taskDate.Format(t), nil
}
