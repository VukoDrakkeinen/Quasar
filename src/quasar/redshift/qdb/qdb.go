package qdb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
)

var dataDir string
var thumbsDir string

func init() {
	luser, _ := user.Current()
	dataDir = filepath.Join(luser.HomeDir, ".local", "share", "quasar")
	thumbsDir = filepath.Join(dataDir, "thumbnails")
	os.MkdirAll(dataDir, os.ModeDir|0755) //TODO: move somewhere better?
	os.Mkdir(thumbsDir, os.ModeDir|0755)
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

func SaveThumbnail(filename string, b []byte) {
	ioutil.WriteFile(filepath.Join(thumbsDir, filename), b, 0644)
}

func GetThumbnailPath(filename string) string {
	return filepath.Join(thumbsDir, filename)
}

type InsertionStmtExecutor interface {
	ExecuteInsertionStmt(stmt *sql.Stmt, additionalArgs ...interface{}) error
}

type QueryStmtExecutor interface {
	ExecuteQueryStmt(stmt *sql.Stmt, additionalArgs ...interface{}) error
}
