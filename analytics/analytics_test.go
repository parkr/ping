package analytics

import (
	"os"
	"testing"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

var db = sqlx.MustConnect("mysql", os.Getenv("PING_DB"))

func init() {
	db.MustExec(`INSERT INTO visits (ip, host, path, user_agent, created_at) VALUES ('127.0.0.1', 'example.org', '/root', 'go test client', NOW())`)
}

func TestVisitorsForPath(t *testing.T) {
	visitors, err := VisitorsForPath(db, "/root")

	if err != nil {
		t.Fatal(err)
	}

	if visitors < 0 {
		t.Error("Visitors should exists")
	}
}

func TestViewsForPath(t *testing.T) {
	views, err := VisitorsForPath(db, "/root")

	if err != nil {
		t.Fatal(err)
	}

	if views < 0 {
		t.Error("Visitors should exists")
	}
}

func TestAllPaths(t *testing.T) {
	paths, err := AllPaths(db)

	if err != nil {
		t.Fatal(err)
	}

	expected := "/root"

	if paths[0] != expected {
		t.Errorf("Got %v want %v", paths[0], expected)
	}
}

func TestAllHosts(t *testing.T) {
	hosts, err := AllHosts(db)

	if err != nil {
		t.Fatal(err)
	}

	expected := "example.org"

	if hosts[0] != expected {
		t.Errorf("Got %v want %v", hosts[0], expected)
	}
}

func TestListDistinctColumnHost(t *testing.T) {
	hosts, err := ListDistinctColumn(db, "host")

	if err != nil {
		t.Fatal(err)
	}

	expected := "example.org"

	if hosts[0] != expected {
		t.Errorf("Got %v want %v", hosts[0], expected)
	}
}

func TestListDistinctColumnPath(t *testing.T) {
	hosts, err := ListDistinctColumn(db, "path")

	if err != nil {
		t.Fatal(err)
	}

	expected := "/root"

	if hosts[0] != expected {
		t.Errorf("Got %v want %v", hosts[0], expected)
	}
}

func TestListDistinctColumnError(t *testing.T) {
	hosts, err := ListDistinctColumn(db, "mehehe")

	if err == nil {
		t.Errorf("Expected an SQL error, got %v", hosts)
	}
}
