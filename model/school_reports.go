package model

import "time"

type SchoolReports struct {
	ID   string    `db:"id"`
	Name string    `db:"name"`
	Date time.Time `db:"date"`
}
