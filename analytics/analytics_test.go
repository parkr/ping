package analytics

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/parkr/ping/database"
)

func initDB() (*sqlx.DB, error) {
	db, err := database.InitializeForTest()
	if err != nil {
		return db, err
	}
	_, err = db.Exec(`
		INSERT INTO visits (ip, host, path, user_agent, created_at) VALUES
		('127.0.0.1', 'example.org', '/root', 'go test client', datetime('now')),
		('127.0.0.1', 'example.org', '/foo', 'go test client', datetime('now'));`)
	return db, err
}

func TestVisitorsForPath(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

	visitors, err := VisitorsForHostPath(db, "example.org", "/root")

	if err != nil {
		t.Fatal(err)
	}

	if visitors != 1 {
		t.Errorf("expected 1 visitor, got: %d", visitors)
	}
}

func TestViewsForHostPath(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

	views, err := ViewsForHostPath(db, "example.org", "/root")

	if err != nil {
		t.Fatal(err)
	}

	if views < 0 {
		t.Error("Visitors should exists")
	}
}

func TestAllPaths(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

	paths, err := AllPaths(db)

	if err != nil {
		t.Fatal(err)
	}

	expected := "/root"

	if paths[0] != expected {
		t.Errorf("Got %v want %v", paths[0], expected)
	}

	expected = "/foo"

	if paths[1] != expected {
		t.Errorf("Got %v want: %v", paths[1], expected)
	}

	if len(paths) > 2 {
		t.Errorf("Got %d paths, want 1", len(paths))
	}
}

func TestAllHosts(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

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
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

	hosts, err := ListDistinctColumn(db, "host")

	if err != nil {
		t.Fatal(err)
	}

	expected := "example.org"

	if hosts[0] != expected {
		t.Errorf("Got %v want %v", hosts[0], expected)
	}

	if len(hosts) > 1 {
		t.Errorf("Got %d hosts, wanted 1", len(hosts))
	}
}

func TestListDistinctColumnPath(t *testing.T) {
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

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
	db, err := initDB()
	if err != nil {
		t.Fatalf("unable to initialize db: %v", err)
	}
	defer db.Close()

	hosts, err := ListDistinctColumn(db, "mehehe")

	if err == nil {
		t.Errorf("Expected an SQL error, got %v", hosts)
	}
}
