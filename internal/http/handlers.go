package http

import (
	"go_final_project/internal/app"
	"go_final_project/internal/date"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gocraft/dbr/v2"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// RegisterHandlers регистрирует обработчики маршрутов
func RegisterHandlers(e *echo.Echo, db app.DBHandler) {
	e.GET("/api/nextdate", NextDateHandler(db))
	e.POST("/api/task", AddTaskHandler(db))
	e.GET("/api/tasks", GetTasksHandler(db))
	e.GET("/api/task", GetTaskByIDHandler(db))
	e.PUT("/api/task", UpdateTaskHandler(db))
	e.POST("/api/task/done", DoneTaskHandler(db))
	e.DELETE("/api/task", DeleteTaskHandler(db))
}

// DeleteTaskHandler обрабатывает запрос на удаление задачи
func DeleteTaskHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.QueryParam("id")
		if err := validateID(id); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		taskID, err := strconv.Atoi(id)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}
		// Удаляем задачу из базы данных
		err = db.DeleteTask(strconv.Itoa(taskID))
		if err != nil {
			logrus.Errorf("Ошибка удаления задачи из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи из базы данных"})
		}

		return c.JSON(http.StatusOK, map[string]string{})
	}
}

func DoneTaskHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := c.QueryParam("id")
		if err := validateID(id); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		task, err := db.GetTaskByID(id)
		if err != nil {
			if err.Error() == "задача не найдена" {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
			}
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задачи"})
		}
		// Проверка правила повторения
		now := time.Now().UTC()
		if task.Repeat == "" {
			// Удаляем одноразовую задачу
			err := db.MarkTaskAsDone(*task, "")
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка удаления задачи"})
			}
		} else {
			// Рассчитываем следующую дату для периодической задачи
			nextDate, err := date.CalculateNextDate(now, task.Date, task.Repeat)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка расчета следующей даты выполнения задачи"})
			}

			// Обновляем задачу с новой датой
			err = db.MarkTaskAsDone(*task, nextDate)
			if err != nil {
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления задачи"})
			}
		}

		return c.JSON(http.StatusOK, map[string]string{})
	}
}

// UpdateTaskHandler обновляет параметры задачи
func UpdateTaskHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		logrus.Debug("Начало обработчика обновления задачи")

		var req app.Models
		if err := c.Bind(&req); err != nil {
			logrus.Errorf("Ошибка десериализации JSON: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка десериализации JSON"})
		}
		// Устанавливаем id в структуру req
		if req.ID == "" {
			logrus.Errorf("Не указан идентификатор задачи")
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Не указан идентификатор"})
		}

		//валидация id
		if err := validateID(req.ID); err != nil {
			logrus.Errorf("Ошибка валидации ID: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		//валидация заголовка
		if err := validateTitle(req.Title); err != nil {
			logrus.Errorf("Ошибка валидации заголовка: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		//валидация даты
		taskDate, err := validateDate(req.Date)
		if err != nil {
			logrus.Errorf("Ошибка валидации даты: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		//валидация даты повтора
		req.Date, err = validateRepeatDate(taskDate, req.Repeat)
		if err != nil {
			logrus.Errorf("Ошибка валидации даты повтора: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		if req.Date == "" || taskDate.Before(time.Now()) {
			req.Date = time.Now().Format("20060102")
		}
		//обновление задачи в базе данных
		rowsAffected, err := db.UpdateTask(req)
		if err != nil {
			logrus.Errorf("Ошибка обновления задачи в базе данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка обновления задачи в базе данных"})
		}
		// была ли задача обновлена
		if rowsAffected == 0 {
			logrus.Warn("Задача не найдена")
			return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
		}
		logrus.Debug("Задача успешно обновлена")
		//возвращаем ответ
		return c.JSON(http.StatusOK, map[string]interface{}{})
	}
}

// GetTaskByIDHandler возвращает параметры задачи по её идентификатору
func GetTaskByIDHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		idStr := c.QueryParam("id")
		if err := validateID(idStr); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid ID"})
		}

		task, err := db.GetTaskByID(strconv.Itoa(id))
		if err != nil {
			if err == dbr.ErrNotFound {
				return c.JSON(http.StatusNotFound, map[string]string{"error": "Задача не найдена"})
			}
			logrus.Errorf("Ошибка получения задачи из базы данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задачи из базы данных"})
		}
		return c.JSON(http.StatusOK, task)
	}
}

// GetTasksHandler возвращает список всех задач
func GetTasksHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		search := c.QueryParam("search")
		var tasks []app.Models
		var err error

		if search != "" {
			// Проверяем, является ли строка поиска датой
			if parsedDate, err := time.Parse(app.UserDateFormat, search); err == nil {
				// Если это дата, то ищем задачи по дате
				tasks, err = db.GetTasksByDate(parsedDate.Format(app.FormatDate))
				if err != nil {
					logrus.Errorf("Ошибка получения задач по дате из базы данных: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задач по дате из базы данных"})
				}
			} else {
				// Если это не дата, то ищем задачи по строке поиска в заголовке или комментарии
				tasks, err = db.GetTasksBySearch(search)
				if err != nil {
					logrus.Errorf("Ошибка получения задач по строке поиска из базы данных: %v", err)
					return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задач по строке поиска из базы данных"})
				}
			}
		} else {
			// Если нет строки поиска, возвращаем все задачи
			tasks, err = db.GetTasks("")
			if err != nil {
				logrus.Errorf("Ошибка получения задач из базы данных: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка получения задач из базы данных"})
			}
		}

		logrus.Infof("Полученные задачи из базы данных: %+v", tasks)

		// Сортировка задач по дате
		sort.Slice(tasks, func(i, j int) bool {
			return tasks[i].Date < tasks[j].Date
		})

		// Ограничение количества возвращаемых задач до 50
		if len(tasks) > 50 {
			tasks = tasks[:50]
		}

		// Проверка на nil и присвоение пустого массива, если tasks == nil
		if tasks == nil {
			tasks = []app.Models{}
		}

		logrus.Infof("Задачи после сортировки и ограничения: %+v", tasks)

		return c.JSON(http.StatusOK, map[string]interface{}{"tasks": tasks})
	}
}

// AddTaskHandler обрабатывает запрос на добавление задачи
func AddTaskHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req app.Models
		if err := c.Bind(&req); err != nil {
			logrus.Errorf("Ошибка десериализации JSON: %v", err)
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Ошибка десериализации JSON"})
		}

		// Проверка обязательного поля Title
		if err := validateTitle(req.Title); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Проверка формата даты
		taskDate, err := validateDate(req.Date)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Если дата меньше сегодняшнего числа и есть правило повторения
		req.Date, err = validateRepeatDate(taskDate, req.Repeat)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
		}

		// Добавление задачи в бд
		id, err := db.AddTask(req)
		if err != nil {
			logrus.Errorf("Ошибка добавления задачи в базу данных: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка добавления задачи в базу данных"})
		}

		logrus.Infof("Задача успешно добавлена с идентификатором %d", id)
		return c.JSON(http.StatusOK, map[string]int64{"id": id})
	}
}

// NextDateHandler обработчик для вычисления следующей даты.
func NextDateHandler(db app.DBHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Получаем параметры запроса
		nowStr := c.QueryParam("now")
		dateStr := c.QueryParam("date")
		repeatStr := c.QueryParam("repeat")

		// Валидируем параметр now
		nowTime, err := validateDate(nowStr)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректная дата now"})
		}

		// // Преобразуем nowTimeString в time.Time
		// nowTime, err := time.Parse(app.FormatDate, nowTimeString)
		// if err != nil {
		// 	return c.JSON(http.StatusBadRequest, map[string]string{"error": "Некорректная дата now"})
		// }

		// Вычисляем следующую дату
		nextDate, err := date.CalculateNextDate(nowTime, dateStr, repeatStr)
		if err != nil {
			logrus.Errorf("Ошибка расчета следующей даты: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Ошибка расчета следующей даты"})
		}

		// Возвращаем результат
		return c.String(http.StatusOK, nextDate)
	}
}
