package app

// Модель
type Models struct {
	ID      string `json:"id"        db:"id"`
	Date    string `json:"date"      db:"date"`
	Title   string `json:"title"     db:"title"`
	Comment string `json:"comment"   db:"comment"`
	Repeat  string `json:"repeat"    db:"repeat"`
}

const (
	UserDateFormat = "02.01.2006"
	FormatDate     = "20060102"
)

type DBHandler interface {
	AddTask(task Models) (int64, error)
	DeleteTask(id string) error
	GetTaskByID(id string) (*Models, error)
	GetTasks(search string) ([]Models, error)
	UpdateTask(task Models) (int64, error)
	MarkTaskAsDone(task Models, nextDate string) error
	GetTasksByDate(date string) ([]Models, error)
	GetTasksBySearch(search string) ([]Models, error)
}
