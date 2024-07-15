package http

import (
	"go_final_project/internal/app"
	"go_final_project/internal/date"
	"net/http"

	"time"

	"github.com/labstack/echo/v4"
)

func validateID(id string) error {
	if id == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Не указан идентификатор")
	}
	return nil
}

func validateTitle(title string) error {
	if title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "Не указан заголовок задачи")
	}
	return nil
}

func validateDate(taskDate string) (string, error) {
	if taskDate == "" {
		return time.Now().Format(app.FormatDate), nil
	}
	parsedDate, err := time.Parse(app.FormatDate, taskDate)
	if err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "Дата представлена в некорректном формате")
	}
	return parsedDate.Format(app.FormatDate), nil
}

func validateRepeatDate(taskDate, repeat string) (string, error) {
	if repeat != "" {
		parsedDate, err := time.Parse(app.FormatDate, taskDate)
		if err != nil {
			return "", echo.NewHTTPError(http.StatusBadRequest, "Дата представлена в некорректном формате")
		}
		if parsedDate.Before(time.Now()) {
			nextDate, err := date.CalculateNextDate(time.Now(), taskDate, repeat)
			if err != nil {
				return "", echo.NewHTTPError(http.StatusBadRequest, "Некорректное правило повторения")
			}
			return nextDate, nil
		}
	}
	return taskDate, nil
}
