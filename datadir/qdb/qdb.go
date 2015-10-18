package qdb

import (
	"database/sql"
	"github.com/VukoDrakkeinen/Quasar/datadir"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

const dbFile = "quasar.db"

var thumbsDir string
var qdb *sql.DB

func init() {
	thumbsDir = filepath.Join(datadir.Path(), "thumbnails")
	os.Mkdir(thumbsDir, os.ModeDir|0755)
}

type QDB struct {
	*sql.DB
}

func (this *QDB) MustPrepare(query string) *sql.Stmt {
	ret, err := this.Prepare(query)
	if err != nil {
		panic(`DB: Prepare(` + strconv.Quote(query) + `): ` + qerr.NewEmbeddedLocated(err).Error())
	}
	return ret
}

func DB() *QDB {
	if qdb == nil {
		var err error //WORKAROUND: syntax analyzer complains
		qdb, err = sql.Open("sqlite3", filepath.Join(datadir.Path(), dbFile))
		if err != nil {
			qlog.Log(qlog.Error, "Opening database failed.", err)
			return nil
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
	if filename == "" {
		return ""
	}
	return filepath.Join(thumbsDir, filename)
}

func ThumbnailExists(filename string) bool {
	if fullpath := GetThumbnailPath(filename); fullpath != "" {
		_, err := os.Lstat(fullpath)
		return !os.IsNotExist(err)
	}
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
