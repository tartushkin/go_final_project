package myDB

import (
	//"go_final_project/internal/app"
	"os"
	"path/filepath"

	"github.com/gocraft/dbr/v2"
	"github.com/sirupsen/logrus"

	_ "modernc.org/sqlite"
)

const DefaultDBFile = "scheduler.db"

func initDB(dbFile string) (*dbr.Session, error) {
	if dbFile == "" {
		appPath, err := os.Executable()
		if err != nil {
			logrus.Fatalf("Ошибка получения пути к исполняемому файлу: %v, err")
			return nil, err
		}
		dbFile = filepath.Join(filepath.Dir(appPath), DefaultDBFile)
	}

	//проверяем наличие файла базы данных
	_, err := os.Stat(dbFile)
	install := os.IsNotExist(err)

	//подключение к базе данных
	conn, err := dbr.Open("sqlite", dbFile, nil)
	if err != nil {
		logrus.Fatalf("Ошибка подключения к базе данных: %v", err)
		return nil, err
	}
	sess := conn.NewSession(nil)
	// Если базы данных не существует, создаем таблицу и индекс
	if install {
		logrus.Info("Файл базы данных не найден, создаем новую базу данных")
		createTableSQL := `
		CREATE TABLE scheduler (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			date CHAR(8) NOT NULL,
			title VARCHAR(128) NOT NULL,
			comment TEXT,
			repeat VARCHAR(128)
		);`
		_, err = sess.Exec(createTableSQL)
		if err != nil {
			logrus.Fatalf("Ошибка создания таблицы: %v", err)
			return nil, err
		}

		createIndexSQL := `CREATE INDEX idx_date ON scheduler (date);`
		_, err = sess.Exec(createIndexSQL)
		if err != nil {
			logrus.Fatalf("Ошибка создания индекса: %v", err)
			return nil, err
		}
		logrus.Info("База данных успешно создана")
	}

	return sess, nil
}
