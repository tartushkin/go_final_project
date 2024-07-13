package myDB

import (
	"os"

	"github.com/gocraft/dbr/v2"
	"github.com/sirupsen/logrus"
	_ "modernc.org/sqlite"
)

const DefaultDBFile = "scheduler.db"

// DB реализует интерфейс app.DBSession
type DB struct {
	*dbr.Session
}

// NewDB создает новую сессию базы данных или подключается к существующей
func NewDB() (*DB, error) {
	if _, err := os.Stat(DefaultDBFile); os.IsNotExist(err) {
		file, err := os.Create(DefaultDBFile)
		if err != nil {
			return nil, err
		}
		file.Close()
	}
	// Если файл не существует, создаем базу данных
	dbConn, err := dbr.Open("sqlite", DefaultDBFile, nil)
	if err != nil {
		logrus.Fatalf("Ошибка подключения к базе данных: %v", err)
		return nil, err
	}

	// Создаем таблицу scheduler
	createTableSQL := `CREATE TABLE IF NOT EXISTS scheduler (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		date CHAR(8) NOT NULL DEFAULT "",
		title VARCHAR(128) NOT NULL DEFAULT "",
		comment TEXT NOT NULL DEFAULT "",
		repeat VARCHAR(128) NOT NULL DEFAULT ""
	);`

	_, err = dbConn.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	// Создаем индекс по полю date для сортировки задач по дате
	createIndexSQL := `CREATE INDEX IF NOT EXISTS idx_scheduler_date ON scheduler(date);`
	_, err = dbConn.Exec(createIndexSQL)
	if err != nil {
		return nil, err
	}

	// Создаем сессию dbr и оборачиваем ее в DB
	sess := dbConn.NewSession(nil)
	return &DB{sess}, nil
}
