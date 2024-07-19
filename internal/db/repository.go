package db

import (
	"fmt"
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

func (db *DB) GetTaskByID(id string) (*app.Models, error) {
	var task app.Models
	_, err := db.Select("*").From("scheduler").Where("id = ?", id).Load(&task)
	if err != nil {
		return nil, err
	}
	if task.ID == "" {
		return nil, dbr.ErrNotFound
	}
	return &task, nil
}

func (db *DB) GetTasks(search string) ([]app.Models, error) {
	var tasks []struct {
		ID      int    `db:"id"`
		Date    string `db:"date"`
		Title   string `db:"title"`
		Comment string `db:"comment"`
		Repeat  string `db:"repeat"`
	}
	query := db.Select("*").From("scheduler")
	if search != "" {
		query = query.Where("title LIKE ?", "%"+search+"%")
	}

	_, err := query.Load(&tasks)
	if err != nil {
		return nil, err
	}

	var result []app.Models
	for _, task := range tasks {
		result = append(result, app.Models{
			ID:      fmt.Sprintf("%d", task.ID),
			Date:    task.Date,
			Title:   task.Title,
			Comment: task.Comment,
			Repeat:  task.Repeat,
		})
	}
	logrus.Infof("Полученные задачи из базы данных: %+v", result)

	return result, nil
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

func (db *DB) DeleteTask(id string) error {
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

// GetTasksByDate возвращает задачи на определенную дату
func (db *DB) GetTasksByDate(date string) ([]app.Models, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE date = ? ORDER BY date LIMIT 50`
	rows, err := db.Query(query, date)
	if err != nil {
		logrus.Errorf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []app.Models
	for rows.Next() {
		var task app.Models
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			logrus.Errorf("Ошибка сканирования строки результата: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

// GetTasksBySearch возвращает задачи по строке поиска
func (db *DB) GetTasksBySearch(search string) ([]app.Models, error) {
	query := `SELECT id, date, title, comment, repeat FROM scheduler WHERE title LIKE ? OR comment LIKE ? ORDER BY date LIMIT 50`
	rows, err := db.Query(query, fmt.Sprintf("%%%s%%", search), fmt.Sprintf("%%%s%%", search))
	if err != nil {
		logrus.Errorf("Ошибка выполнения запроса к базе данных: %v", err)
		return nil, err
	}
	defer rows.Close()

	var tasks []app.Models
	for rows.Next() {
		var task app.Models
		if err := rows.Scan(&task.ID, &task.Date, &task.Title, &task.Comment, &task.Repeat); err != nil {
			logrus.Errorf("Ошибка сканирования строки результата: %v", err)
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
