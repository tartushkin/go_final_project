package db

import (
	"go_final_project/internal/app"
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

func (db *DB) AddTask(task app.Models) (int64, error) {
	result, err := db.InsertInto("scheduler").
		Columns("date", "title", "comment", "repeat").
		Values(task.Date, task.Title, task.Comment, task.Repeat).
		Exec()
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (db *DB) GetTaskByID(id int) (*app.Models, error) {
	var task app.Models
	_, err := db.Select("*").From("scheduler").Where("id = ?", id).Load(&task)
	if err != nil {
		return nil, err
	}
	if task.ID == 0 {
		return nil, dbr.ErrNotFound
	}
	return &task, nil
}

func (db *DB) GetTasks(search string) ([]app.Models, error) {
	var tasks []app.Models
	query := db.Select("*").From("scheduler")
	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	_, err := query.Load(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (db *DB) UpdateTask(task app.Models) (int64, error) {
	result, err := db.Update("scheduler").
		Set("date", task.Date).
		Set("title", task.Title).
		Set("comment", task.Comment).
		Set("repeat", task.Repeat).
		Where("id = ?", task.ID).
		Exec()
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}

func (db *DB) DeleteTask(id int) error {
	_, err := db.DeleteFrom("scheduler").Where("id = ?", id).Exec()
	return err
}

func (db *DB) MarkTaskAsDone(task app.Models, nextDate string) error {
	if nextDate != "" {
		_, err := db.Update("scheduler").
			Set("date", nextDate).
			Where("id = ?", task.ID).
			Exec()
		if err != nil {
			return err
		}
	} else {
		_, err := db.DeleteFrom("scheduler").Where("id = ?", task.ID).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}
