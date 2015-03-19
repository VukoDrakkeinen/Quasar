package redshift

import (
	"fmt"
	"log"
	"quasar/qutils/qerr"
	"quasar/redshift/idsdict"
	"quasar/redshift/qdb"
	"strings"
)

var (
	idsSchema = `
	CREATE TABLE IF NOT EXISTS langs(
		id INTEGER PRIMARY KEY,
		lang TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS scanlators(
		id INTEGER PRIMARY KEY,
		scanlator TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS altTitles(
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS authors(
		id INTEGER PRIMARY KEY,
		author TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS artists(
		id INTEGER PRIMARY KEY,
		artist TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS genres(
		id INTEGER PRIMARY KEY,
		genre TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS tags(
		id INTEGER PRIMARY KEY,
		tag TEXT UNIQUE NOT NULL
	);
	`
	idsInsertionPreCmd = `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	idsQueryPreCmd     = `SELECT $colName FROM $tableName;` //TODO?: use placeholders?

)

type ComicList []*Comic

func CreateDB(db *qdb.QDB) (err error) {
	transaction, _ := db.Begin()
	defer transaction.Rollback()
	_, err = transaction.Exec(idsSchema)
	if err != nil {
		return qerr.NewLocated(err)
	}
	_, err = transaction.Exec(SQLComicSchema())
	if err != nil {
		return qerr.NewLocated(err)
	}
	transaction.Commit()
	return
}

func (this ComicList) SaveToDB() { //TODO: write a unit test
	log.SetFlags(log.Lshortfile | log.Ltime)
	db := qdb.DB()
	if db == nil {
		fmt.Println("Database handle is nil! Aborting save.") //TODO: proper logging
		//panic()	//TODO?
	}
	err := CreateDB(db)
	if err != nil {
		fmt.Println(err) //TODO: proper logging
		//panic()	//TODO?
		return
	}

	type tuple struct {
		dict qdb.InsertionStmtExecutor
		name string
	}
	for _, tuple := range []tuple{ //TODO?: global state, hmm
		{&idsdict.Langs, "lang"},
		{&idsdict.Scanlators, "scanlator"},
		{&idsdict.Authors, "author"},
		{&idsdict.Artists, "artist"},
		{&idsdict.ComicGenres, "genre"},
		{&idsdict.ComicTags, "tag"},
	} {
		transaction, _ := db.Begin()
		defer transaction.Rollback()
		rep := strings.NewReplacer("$tableName", tuple.name+"s", "$colName", tuple.name)
		idsInsertionStmt, _ := transaction.Prepare(rep.Replace(idsInsertionPreCmd))
		err = tuple.dict.ExecuteInsertionStmt(idsInsertionStmt)
		if err != nil {
			fmt.Println(err) //TODO: proper logging
			return
		}
		transaction.Commit()
	}

	dbStmts := SQLComicInsertStmts(db)
	defer dbStmts.Close()
	for _, comic := range this {
		transaction, _ := db.Begin()
		stmts := dbStmts.ToTransactionSpecific(transaction)

		err := comic.SQLInsert(stmts)
		if err != nil { // statements are closed by Commit() or Rollback()
			transaction.Rollback()
			fmt.Println(err)
		} else {
			transaction.Commit()
		}
	}
}

func LoadComicList() (list ComicList, err error) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	db := qdb.DB()
	CreateDB(db)
	transaction, _ := db.Begin()
	defer transaction.Rollback()

	type tuple struct {
		dict qdb.QueryStmtExecutor
		name string
	}
	for _, tuple := range []tuple{ //TODO?: dicts as function arguments? (global state side effects are not nice)
		{&idsdict.Langs, "lang"},
		{&idsdict.Scanlators, "scanlator"},
		{&idsdict.Authors, "author"},
		{&idsdict.Artists, "artist"},
		{&idsdict.ComicGenres, "genre"},
		{&idsdict.ComicTags, "tag"},
	} {
		rep := strings.NewReplacer("$tableName", tuple.name+"s", "$colName", tuple.name)
		idsQueryStmt, _ := transaction.Prepare(rep.Replace(idsQueryPreCmd))
		err := tuple.dict.ExecuteQueryStmt(idsQueryStmt)
		if err != nil {
			return list, err
		}
		idsQueryStmt.Close()
	}

	dbStmts := SQLComicQueryStmts(db)
	defer dbStmts.Close()
	stmts := dbStmts.ToTransactionSpecific(transaction)
	comicRows, err := stmts[comicsQuery].Query()
	if err != nil {
		return list, qerr.NewLocated(err)
	}
	for comicRows.Next() {
		comic, err := SQLComicQuery(comicRows, stmts)
		if err != nil {
			return list, err
		}
		list = append(list, comic)
	}

	transaction.Commit()
	return
}
