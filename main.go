package main

import (
	"go_final_project/internal/http"
	"go_final_project/internal/myDB"
	"os"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
)

func main() {
	//инициализация экземпляра echo и middleware
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	//инициализация логов
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)
	logrus.SetLevel(logrus.InfoLevel)

	// Получаем значение порта из переменной окружения TODO_PORT
	port := os.Getenv("TODO_PORT")
	if port == "" {
		// Если переменная окружения нет, используем порт по умолчанию 7540
		port = ":7540"
	} else {
		if _, err := strconv.Atoi(port); err == nil {
			port = ":" + port
		}
	}

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		// Use default path if TODO_DBFILE is not set
		dbFile = "./scheduler.db"
	}

	//инициализация репо для работы c базой
	db, err := myDB.NewDB()
	if err != nil {
		logger.Fatalf("Ошбка инициализации базы данных: %v", err)
	}
	defer db.Close()

	// Настройка статических файлов
	webDir := "./web"
	e.Static("/", webDir)

	// Регистрация маршрутов
	http.RegisterHandlers(e, db)

	//Запуск сервера в горутине
	go func() {
		if err := e.Start(port); err != nil {
			logger.Fatalf("Ошибка запуска сервера %v", err)
		}
	}()
	logger.Info("Cервер успешно запущен на порту " + port)
	//блокируем горутину
	select {}
}
