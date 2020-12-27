package database

import (
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

const (
	// This is the format for a SQL Datetime Literal.
	SQLDateTimeFormat = "2006-01-02 15:04:05"

	schema = `CREATE TABLE visits (
		id integer NOT NULL PRIMARY KEY AUTOINCREMENT,
    	ip varchar(255) NOT NULL,
    	host text NOT NULL,
      	user_agent text NOT NULL,
    	path text NOT NULL,
		created_at datetime NOT NULL
	);`
	checkIfSchemaExists = `SELECT COUNT(*) as does_exist FROM sqlite_master WHERE type='table' AND name='visits';`
	insertVisit         = `INSERT INTO visits (ip, host, path, user_agent, created_at) VALUES (:ip, :host, :path, :user_agent, :created_at)`
)

type TableCheck struct {
	DoesExist int `db:"does_exist"`
}

func Initialize() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", os.Getenv("PING_DB"))
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return db, err
	}
	var check TableCheck
	if err := db.Get(&check, checkIfSchemaExists); err != nil {
		return db, err
	}
	if check.DoesExist < 1 {
		if _, err := db.Exec(schema); err != nil {
			return db, err
		}
	}
	return db, nil
}

type Visit struct {
	IP        string `db:"ip"`
	Host      string `db:"host"`
	Path      string `db:"path"`
	UserAgent string `db:"user_agent"`
	CreatedAt string `db:"created_at"`
}

func (v *Visit) String() string {
	return fmt.Sprintf("<%s | %s requested %s%s @ %s>", v.CreatedAt, v.IP, v.Host, v.Path, v.UserAgent)
}

func (v *Visit) Save(db *sqlx.DB) error {
	_, err := db.NamedExec(insertVisit, v)
	return err
}
