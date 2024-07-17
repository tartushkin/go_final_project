package http

import (
	"go_final_project/internal/app"
	"go_final_project/internal/date"
	"net/http"

	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
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

func validateDate(taskDate string) (string, error) {
	logrus.Debugf("Проверка даты: %s", taskDate)
	if taskDate == "" {
		logrus.Debug("Дата пустая, используется текущая дата")
		return time.Now().Format(app.FormatDate), nil
	}
	parsedDate, err := time.Parse(app.FormatDate, taskDate)
	if err != nil {
		logrus.Debug("Дата в некорректном формате")
		return "", echo.NewHTTPError(http.StatusBadRequest, "Дата представлена в некорректном формате")
	}
	logrus.Debugf("Дата валидна: %s", parsedDate.Format(app.FormatDate))
	return parsedDate.Format(app.FormatDate), nil
}

func validateRepeatDate(taskDate, repeat string) (string, error) {
	if repeat != "" {
		parsedDate, err := time.Parse(app.FormatDate, taskDate)
		if err != nil {
			logrus.Debug("Дата в некорректном формате для правила повтора")
			return "", echo.NewHTTPError(http.StatusBadRequest, "Дата представлена в некорректном формате")
		}
		now := time.Now().Truncate(24 * time.Hour)
		if parsedDate.Before(now) {
			nextDate, err := date.CalculateNextDate(now, taskDate, repeat)
			if err != nil {
				logrus.Debug("Некорректное правило повторения")
				return "", echo.NewHTTPError(http.StatusBadRequest, "Некорректное правило повторения")
			}
			logrus.Debugf("Дата повтора валидна: %s", nextDate)
			return nextDate, nil
		}
	}
	logrus.Debugf("Дата повтора валидна: %s", taskDate)
	return taskDate, nil
}
