package app

import (
	"github.com/gocraft/dbr/v2"
)

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

type DBSession interface {
	Select(columns ...string) *dbr.SelectStmt
	InsertInto(table string) *dbr.InsertStmt
	Update(table string) *dbr.UpdateStmt
	DeleteFrom(table string) *dbr.DeleteStmt
}
