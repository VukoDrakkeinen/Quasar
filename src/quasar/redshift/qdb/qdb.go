package qdb

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
)

const dbFile = "quasar.db"

var dataDir string
var thumbsDir string
var qdb *sql.DB

func init() {
	luser, _ := user.Current()
	dataDir = filepath.Join(luser.HomeDir, ".local", "share", "quasar")
	thumbsDir = filepath.Join(dataDir, "thumbnails")
	os.MkdirAll(dataDir, os.ModeDir|0755) //TODO: move somewhere better?
	os.Mkdir(thumbsDir, os.ModeDir|0755)
}

type QDB struct {
	*sql.DB
}

func (this *QDB) MustPrepare(query string) *sql.Stmt {
	ret, err := this.Prepare(query)
	if err != nil {
		panic(`DB: Prepare(` + strconv.Quote(query) + `): ` + err.Error())
	}
	return ret
}

func DB() *QDB {
	if qdb == nil {
		var err error //WORKAROUND: syntax analyzer complains
		qdb, err = sql.Open("sqlite3", filepath.Join(dataDir, dbFile))
		if err != nil {
			//TODO: log error
		}
		qdb.Exec(`PRAGMA foreign_keys = ON;`) //enable foreign keys
	}
	return &QDB{qdb}
}

type StmtGroup map[string]*sql.Stmt

func (this StmtGroup) ToTransactionSpecific(transaction *sql.Tx) StmtGroup {
	specific := StmtGroup(make(map[string]*sql.Stmt, len(this)))
	for k, v := range this {
		specific[k] = transaction.Stmt(v)
	}
	return specific
}

func (this StmtGroup) Close() {
	for _, v := range this {
		v.Close()
	}
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

type SQLInsertable interface {
	SQLInsert(stmt StmtGroup, additionalArgs ...interface{}) error
}
