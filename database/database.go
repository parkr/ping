package database

import (
	"fmt"

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
	selectVisit         = `SELECT ip, host, path, user_agent, created_at FROM visits WHERE id = ?`
)

type TableCheck struct {
	DoesExist int `db:"does_exist"`
}

// InitializeForTest creates an in-memory SQL database for tests only.
func InitializeForTest() (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", "") // An empty string appears to create a one-off, in-memory database.
	if err != nil {
		return db, err
	}
	_, err = db.Exec(schema)
	return db, err
}

func Initialize(connection string) (*sqlx.DB, error) {
	db, err := sqlx.Connect("sqlite3", connection)
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

func Get(db *sqlx.DB, id int) (Visit, error) {
	row := db.QueryRow(selectVisit, id)
	if row.Err() != nil {
		return Visit{}, row.Err()
	}
	visit := Visit{}
	err := row.Scan(&visit.IP, &visit.Host, &visit.Path, &visit.UserAgent, &visit.CreatedAt)
	return visit, err
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
