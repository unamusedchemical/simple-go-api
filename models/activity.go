package models

import (
	"database/sql"
	"time"
)

type Activity struct {
	Id       int64
	Title    string
	Body     string
	OpenedOn time.Time
	ClosedOn sql.NullTime
	Due      sql.NullTime
	UserId   uint
	GroupId  sql.NullInt64
}
