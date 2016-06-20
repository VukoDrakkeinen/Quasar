package qdb

import (
	"database/sql"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	//"sync"
	//"sync/atomic"

	"github.com/VukoDrakkeinen/Quasar/datadir"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	_ "github.com/mattn/go-sqlite3"
)

const dbFile = "quasar.db"

var (
	thumbsDir string
	db        *sql.DB
	//stmtTypeCounter uint32

	futureStatements []futureStmt
)

func init() {
	thumbsDir = filepath.Join(datadir.Path(), "thumbnails")
	os.Mkdir(thumbsDir, os.ModeDir|0755)
}

type futureStmt struct {
	assignToVar **sql.Stmt
	source      string
}

func PrepareStmt(assignToVar **sql.Stmt, source string) {
	if db == nil {
		futureStatements = append(futureStatements, futureStmt{assignToVar, source})
	} else {
		panic("It is too late to prepare!") //todo: better message
	}
}

//type StmtId uint32
//
//func NewStmtId() StmtId {
//	return StmtId(atomic.AddUint32(&stmtTypeCounter, 1))
//}

type QDB struct {
	*sql.DB
}

func (this *QDB) MustPrepare(query string) *sql.Stmt {
	ret, err := this.Prepare(query)
	if err != nil {
		panic(`DB: MustPrepare(` + strconv.Quote(query) + `): ` + qerr.NewEmbeddedLocated(err).Error())
	}
	return ret
}

func DB() *QDB { //todo: de-singleton
	if db == nil {
		var err error
		/*ci_crt := func(identity int64) (corrected int64) {
			return identity &^ 0xFF
		}
		ci_ver := func(identity int64) (version int64) {
			return int64(byte(identity))
		}
		sql.Register("sqlite3_quasar", &sqlite3.SQLiteDriver{
			ConnectHook: func(conn *sqlite3.SQLiteConn) error {
				if err := conn.RegisterFunc("ci_crt", ci_crt, true); err != nil {
					return err
				}
				if err := conn.RegisterFunc("ci_ver", ci_ver, true); err != nil {
					return err
				}
				return nil
			},
		})//*/
		db, err = sql.Open("sqlite3", filepath.Join(datadir.Path(), dbFile)) //todo: sql.Open() swallows errors from driver.Open(); is it intentional? report it?
		if err != nil {
			qlog.Log(qlog.Error, "Opening database failed.", err)
			return nil
		}
		db.Exec(`PRAGMA foreign_keys = ON;`) //enable foreign keys
		qdb := &QDB{db}
		for _, fStmt := range futureStatements {
			*fStmt.assignToVar = qdb.MustPrepare(fStmt.source)
		}
	}
	return &QDB{db}
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
	return false
}

type InsertionStmtExecutor interface {
	ExecuteInsertionStmt(stmt *sql.Stmt, additionalArgs ...interface{}) error
}

type QueryStmtExecutor interface {
	ExecuteQueryStmt(stmt *sql.Stmt, additionalArgs ...interface{}) error
}
