package qdb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"os/user"
	"path/filepath"
)

var dataDir string

func init() {
	luser, _ := user.Current()
	dataDir = filepath.Join(luser.HomeDir, ".local", "share", "quasar")
	os.MkdirAll(dataDir, os.ModeDir|0755) //TODO: move somewhere better?
}

const dbFile = "quasar.db"

var qdb *sql.DB

func DB() *sql.DB {
	if qdb == nil {
		var err error //WORKAROUND: syntax analyzer complains
		qdb, err = sql.Open("sqlite3", filepath.Join(dataDir, dbFile))
		if err != nil {
			//TODO: log error
		}
	}
	return qdb
}

type DBStatementExecutor interface {
	ExecuteDBStatement(stmt *sql.Stmt, additionalArgs ...interface{}) error
}
