package http

import (
	"go_final_project/internal/app"
	"go_final_project/internal/logic"
	"net/http"
	"sort"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// RegisterHandlers регистрирует обработчики маршрутов
func RegisterHandlers(e *echo.Echo, db app.DBSession) {
	e.GET("/api/nextdate", NextDateHandler(db))
	e.POST("/api/task", AddTaskHandler(db))
	e.GET("/api/tasks", GetTasksHandler(db))
	e.GET("/api/task/:id", GetTaskByIDHandler(db))
	e.PUT("/api/task", UpdateTaskHandler(db))
	e.POST("/api/task/done", DoneTaskHandler(db))
	e.DELETE("/api/task", DeleteTaskHandler(db))
}

// DeleteTaskHandler обрабатывает запрос на удаление задачи
func DeleteTaskHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		}

		// Удаляем задачу из базы данных
		_, err := db.DeleteFrom("scheduler").Where("id = ?", id).Exec()
		if err != nil {
			logrus.Errorf("Ошибка удаления задачи из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи из базы данных"})
		}

		return c.JSON(http.StatusOK, map[string]string{})
	}
}

// DoneTaskHandler обрабатывает запрос на отметку задачи как выполненной
func DoneTaskHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.QueryParam("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		}

		var task app.Models
		_, err := db.Select("*").From("scheduler").Where("id = ?", id).Load(&task)
		if err != nil {
			logrus.Errorf("Ошибка получения задачи из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Задача не найдена"})
		}

		// Проверяем, что задача с указанным идентификатором существует
		if task.ID == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}

		// Для периодической задачи обновляем дату следующего выполнения
		if task.Repeat != "" {
			// Преобразуем строку task.Date в тип time.Time
			taskDate, err := time.Parse(app.FormatDate, task.Date)
			if err != nil {
				logrus.Errorf("Ошибка парсинга даты задачи: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка парсинга даты задачи"})
			}

			nextDate, err := logic.CalculateNextDate(taskDate, task.Date, task.Repeat)
			if err != nil {
				logrus.Errorf("Ошибка расчета следующей даты выполнения задачи: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка расчета следующей даты выполнения задачи"})
			}
			// Обновляем задачу в базе данных
			_, err = db.Update("scheduler").
				Set("date", nextDate). // Используем новую дату
				Where("id = ?", id).
				Exec()
			if err != nil {
				logrus.Errorf("Ошибка обновления задачи в базе данных: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления задачи в базе данных"})
			}
		} else {
			// Для одноразовой задачи удаляем её из базы данных
			_, err := db.DeleteFrom("scheduler").Where("id = ?", id).Exec()
			if err != nil {
				logrus.Errorf("Ошибка удаления задачи из базы данных: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи из базы данных"})
			}
		}

		return c.JSON(http.StatusOK, map[string]string{})
	}
}

// UpdateTaskHandler обновляет параметры задачи
func UpdateTaskHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req app.Models
		if err := c.Bind(&req); err != nil {
			logrus.Errorf("Ошибка десериализации JSON: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка десериализации JSON"})
		}

		if req.ID == 0 {
			logrus.Errorf("Не указан идентификатор задачи")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор задачи"})
		}

		if req.Title == "" {
			logrus.Errorf("Не указан заголовок задачи")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		}

		var taskDate time.Time
		var err error
		if req.Date == "" {
			taskDate = time.Now()
			req.Date = taskDate.Format(app.FormatDate)
		} else {
			taskDate, err = time.Parse(app.FormatDate, req.Date)
			if err != nil {
				logrus.Errorf("Дата представлена в некорректном формате: %v", err)
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Дата представлена в некорректном формате"})
			}
		}

		if taskDate.Before(time.Now()) && req.Repeat != "" {
			req.Date, err = logic.CalculateNextDate(time.Now(), req.Date, req.Repeat)
			if err != nil {
				logrus.Errorf("Некорректное правило повторения: %v", err)
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректное правило повторения"})
			}
		}

		result, err := db.Update("scheduler").
			Set("date", req.Date).
			Set("title", req.Title).
			Set("comment", req.Comment).
			Set("repeat", req.Repeat).
			Where("id = ?", req.ID).
			Exec()
		if err != nil {
			logrus.Errorf("Ошибка обновления задачи в базе данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления задачи в базе данных"})
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			logrus.Errorf("Ошибка получения количества обновленных строк: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения количества обновленных строк"})
		}

		if rowsAffected == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{})
	}
}

// GetTaskByIDHandler возвращает параметры задачи по её идентификатору
func GetTaskByIDHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.Param("id")
		if id == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		}

		var task app.Models
		_, err := db.Select("*").From("scheduler").Where("id = ?", id).Load(&task)
		if err != nil {
			logrus.Errorf("Ошибка получения задачи из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задачи из базы данных"})
		}

		if task.ID == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}

		return c.JSON(http.StatusOK, task)
	}
}

// GetTasksHandler возвращает список всех задач
func GetTasksHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		var tasks []app.Models
		query := db.Select("*").From("scheduler")

		// Обработка параметра поиска
		search := c.QueryParam("search")
		if search != "" {
			// Проверяем, является ли строка поиска - датой
			if date, err := time.Parse("02.01.2006", search); err == nil {
				query = query.Where("date = ?", date.Format("20060102"))
			} else {
				// Иначе ищем по названию и комменту
				query = query.Where("title LIKE ? OR comment LIKE ?", "%"+search+"%", "%"+search+"%")
			}
		}

		// Выгрузка задач из бд
		_, err := query.Load(&tasks)
		if err != nil {
			logrus.Errorf("Ошибка получения задач из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задач из базы данных"})
		}

		// Сортировка задач по дате
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Date < tasks[j].Date
		})

		// Ограничение количества возвращаемых задач
		if len(tasks) > 50 {
			tasks = tasks[:50]
		}

		if tasks == nil {
			tasks = []app.Models{}
		}

		return c.JSON(http.StatusOK, map[string]interface{}{"tasks": tasks})
	}
}

// AddTaskHandler обрабатывает запрос на добавление задачи
func AddTaskHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req app.Models
		if err := c.Bind(&req); err != nil {
			logrus.Errorf("Ошибка десериализации JSON: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка десериализации JSON"})
		}

		// Проверка обязательного поля Title
		if req.Title == "" {
			logrus.Errorf("Не указан заголовок задачи")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан заголовок задачи"})
		}

		// Проверка формата даты
		var taskDate time.Time
		var err error
		if req.Date == "" {
			taskDate = time.Now()
			req.Date = taskDate.Format(app.FormatDate)
		} else {
			taskDate, err = time.Parse(app.FormatDate, req.Date)
			if err != nil {
				logrus.Errorf("Дата представлена в некорректном формате: %v", err)
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Дата представлена в некорректном формате"})
			}
		}

		// Если дата меньше сегодняшнего числа и есть правило повторения
		if taskDate.Before(time.Now()) && req.Repeat != "" {
			req.Date, err = logic.CalculateNextDate(time.Now(), req.Date, req.Repeat)
			if err != nil {
				logrus.Errorf("Некорректное правило повторения: %v", err)
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректное правило повторения"})
			}
		}

		// Добавление задачи в бд
		result, err := db.InsertInto("scheduler").
			Columns("date", "title", "comment", "repeat").
			Values(req.Date, req.Title, req.Comment, req.Repeat).
			Exec()
		if err != nil {
			logrus.Errorf("Ошибка добавления задачи в базу данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка добавления задачи в базу данных"})
		}

		id, err := result.LastInsertId()
		if err != nil {
			logrus.Errorf("Ошибка получения идентификатора задачи: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения идентификатора задачи"})
		}

		logrus.Infof("Задача успешно добавлена с идентификатором %d", id)
		return c.JSON(http.StatusOK, map[string]int64{"id": id})
	}
}

// NextDateHandler обработчик для вычисления следующей даты.
func NextDateHandler(db app.DBSession) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Получаем параметры запроса
		nowStr := c.QueryParam("now")
		dateStr := c.QueryParam("date")
		repeatStr := c.QueryParam("repeat")

		// Парсим параметр now
		now, err := time.Parse("20060102", nowStr)
		if err != nil {
			return c.String(http.StatusBadRequest, "некорректная дата now")
		}

		// Вычисляем следующую дату
		nextDate, err := logic.CalculateNextDate(now, dateStr, repeatStr)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		// Возвращаем результат
		return c.String(http.StatusOK, nextDate)
	}
}
