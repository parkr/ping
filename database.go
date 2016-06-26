package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const insertVisit = `INSERT INTO visits (ip, host, path, user_agent, created_at) VALUES (:ip, :host, :path, :user_agent, :created_at)`

var db *sqlx.DB

func init() {
	db = sqlx.MustConnect("mysql", os.Getenv("PING_DB"))
	db.Ping()
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

func (v *Visit) Save() error {
	_, err := db.NamedExec(insertVisit, v)
	return err
}
