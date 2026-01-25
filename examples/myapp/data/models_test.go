package data

import (
	"fmt"
	"os"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	db2 "github.com/upper/db/v4"
)

func TestNew(t *testing.T) {
	fakeDB, _, _ := sqlmock.New()
	defer fakeDB.Close()

	_ = os.Setenv("DATABASE_TYPE", "postgres")
	m := New(fakeDB)
	if fmt.Sprintf("%T", m) != "*data.Model" {
		t.Errorf("Expected *data.Model, got %T", m)
	}

	_ = os.Setenv("DATABASE_TYPE", "mysql")
	m = New(fakeDB)
	if fmt.Sprintf("%T", m) != "*data.Model" {
		t.Errorf("Expected *data.Model, got %T", m)
	}
}

func TestGetInsertID(t *testing.T) {
	var id db2.ID
	id = int64(1)
	if getInsertID(id) != 1 {
		t.Errorf("Expected 1, got %d", getInsertID(id))
	}

}
