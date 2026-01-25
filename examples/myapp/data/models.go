package data

import (
	"database/sql"
	"os"
	"reflect"

	db2 "github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/mysql"
	"github.com/upper/db/v4/adapter/postgresql"
)

var db *sql.DB
var upper db2.Session

type Model struct {
	// any models inserted here (and in the New function)
	// are easily accessible throughout the entire application
	Users  User
	Tokens Token
}

func New(databasePool *sql.DB) *Model {
	db = databasePool

	switch os.Getenv("DATABASE_TYPE") {
	case "mysql", "mariadb":
		upper, _ = mysql.New(databasePool)
	default:
		upper, _ = postgresql.New(databasePool)
	}

	return &Model{
		Users:  User{},
		Tokens: Token{},
	}
}

func getInsertID(i db2.ID) int {
	idType := reflect.TypeOf(i)
	if idType.Kind() == reflect.Int64 {
		return int(i.(int64))
	}
	return i.(int)
}
