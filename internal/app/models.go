package app

// Модель
type Models struct {
	ID      int    `json:"id"        db:"id"`
	Date    string `json:"date"      db:"date"`
	Title   string `json:"title"     db:"title"`
	Comment string `json:"comment"   db:"comment"`
	Repeat  string `json:"repeat"    db:"repeat"`
}
