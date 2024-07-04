package main

import (
	//"go_final_project/tests"
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

	// Получаем значение порта из переменной окружения TODO_PORT
	port := os.Getenv("TODO_PORT")
	if port == "" {
		// Если переменная окружения не установлена, используем порт по умолчанию 7540
		port = ":7540"
	} else {
		// Добавляем двоеточие
		if _, err := strconv.Atoi(port); err == nil {
			port = ":" + port
		}
	}

	//инициализация репо для раюоты c базой

	// Настройка статических файлов
	webDir := "./web"
	e.Static("/", webDir)

	//Запуск сервера в горутине
	go func() {
		if err := e.Start(port); err != nil {
			logger.Fatalf("Ошибка запуска сервера %v", err)
		}
	}()
	logger.Info("Cервер успешно запущен на порту :7540")
	//блокирум горутину
	select {}
}
