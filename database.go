package main

import (
	"fmt"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	schema = `CREATE TABLE visits (
		id int(11) NOT NULL AUTO_INCREMENT,
    	ip varchar(255) NOT NULL,
    	host text NOT NULL,
    	path text NOT NULL,
		created_at datetime NOT NULL,
		PRIMARY KEY (id)
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`
	checkIfSchemaExists = `SELECT COUNT(*) as does_exist FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'visits'`
	insertVisit = `INSERT INTO visits (ip, host, path, created_at) VALUES (:ip, :host, :path, :created_at)`
)

var db *sqlx.DB

type TableCheck struct {
	DoesExist int `db:"does_exist"`
}

func init() {
	db = sqlx.MustConnect("mysql", os.Getenv("PING_DB"))
	db.Ping()
	var check TableCheck
	err := db.Get(&check, checkIfSchemaExists)
	if err != nil {
		panic(err)
	}
	if check.DoesExist < 1 {
		db.MustExec(schema)
	}
}

type Visit struct {
	IP        string `db:"ip"`
	Host      string `db:"host"`
	Path      string `db:"path"`
	CreatedAt string `db:"created_at"`
}

func (v *Visit) String() string {
	return fmt.Sprintf("<%s | %s requested %s%s>", v.CreatedAt, v.IP, v.Host, v.Path)
}

func (v *Visit) Save() error {
	_, err := db.NamedExec(insertVisit, v)
	return err
}
