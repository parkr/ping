package main

import (
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	schema = `CREATE TABLE visits (
		id int(11) NOT NULL AUTO_INCREMENT,
    	ip varchar(255) NOT NULL,
    	host text NOT NULL,
      user_agent text NOT NULL,
    	path text NOT NULL,
		created_at datetime NOT NULL,
		PRIMARY KEY (id)
	) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8;`
	checkIfSchemaExists = `SELECT COUNT(*) as does_exist FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'visits'`
)

type TableCheck struct {
	DoesExist int `db:"does_exist"`
}

func VerifySchema() {
	db := sqlx.MustConnect("mysql", os.Getenv("PING_DB"))
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

func main() {
	VerifySchema()
}
