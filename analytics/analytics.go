package analytics

import (
	"fmt"
	"github.com/jmoiron/sqlx"
)

const (
	// Count the number of distinct IP addresses which have visitied the path.
	QueryVisitorsPerPath = `SELECT COUNT(distinct ip) FROM visits WHERE path = ?;`
	// Count the number of entries with the given path.
	QueryVisitsPerPath = `SELECT COUNT(id) FROM visits WHERE path = ?;`

	// List all the distinct paths in the database.
	QueryAllPaths = `SELECT DISTINCT path FROM visits;`
	// List all the distinct hosts in the database.
	QueryAllHosts = `SELECT DISTINCT host FROM visits;`
)

// Fetch a count of all the visitors for the given path. This is done by
// counting the distinct IP addresses which have visited the path.
func VisitorsForPath(db *sqlx.DB, path string) (count int, err error) {
	err = db.Get(&count, QueryVisitorsPerPath, path)
	return count, err
}

// Fetch a count of all the views of the path.
func ViewsForPath(db *sqlx.DB, path string) (count int, err error) {
	err = db.Get(&count, QueryVisitsPerPath, path)
	return count, err
}

// Fetch all the paths in the database.
func AllPaths(db *sqlx.DB) (paths []string, err error) {
	err = db.Select(&paths, QueryAllPaths)
	return paths, err
}

// Fetch all the hosts in the database.
func AllHosts(db *sqlx.DB) (hosts []string, err error) {
	err = db.Select(&hosts, QueryAllHosts)
	return hosts, err
}

// Fetch the distinct entries of an arbitrary column in the database.
func ListDistinctColumn(db *sqlx.DB, col string) (entries []string, err error) {
	query := fmt.Sprintf("SELECT DISTINCT %s FROM visits;", col)
	fmt.Println(query)
	err = db.Select(&entries, query)
	return entries, err
}
