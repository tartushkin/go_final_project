package app

// Модель
type Models struct {
	ID      int    `json:"id"        db:"id"`
	Date    string `json:"date"      db:"date"`
	Title   string `json:"title"     db:"title"`
	Comment string `json:"comment"   db:"comment"`
	Repeat  string `json:"repeat"    db:"repeat"`
}

const (
	FormatDate = "20060102"
)

type DBHandler interface {
	AddTask(task Models) (int64, error)
	DeleteTask(id int) error
	GetTaskByID(id int) (*Models, error)
	GetTasks(search string) ([]Models, error)
	UpdateTask(task Models) (int64, error)
	MarkTaskAsDone(task Models, nextDate string) error
}
