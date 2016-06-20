package core

import (
	"database/sql"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

var (
	linksInsertion *sql.Stmt
	linkDeletion   *sql.Stmt
	sourcesQuery   *sql.Stmt
)

func init() {
	qdb.PrepareStmt(&linksInsertion, `
		INSERT OR REPLACE INTO sourceLinks(id, comicId, sourceId, url, markAsRead)
		VALUES((SELECT id FROM sourceLinks WHERE id = ?), ?, ?, ?, ?);`) //todo: remove optional id?
	qdb.PrepareStmt(&linkDeletion, `DELETE FROM sourceLinks WHERE id = ?`)
	qdb.PrepareStmt(&sourcesQuery, `SELECT id, sourceId, url, markAsRead FROM sourceLinks WHERE comicId = ?;`)
}

func SQLSourceLinkSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS sourceLinks(
		id INTEGER PRIMARY KEY,
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		sourceId TEXT NOT NULL,
		url TEXT NOT NULL,
		markAsRead INTEGER NOT NULL --bool
	);
	CREATE INDEX IF NOT EXISTS sourceLinks_cid_idx On sourceLinks(comicId);`
}

type SourceLink struct {
	SourceId   SourceId
	URL        string
	MarkAsRead bool

	sqlId int64
}

func (this *SourceLink) sqlInsert(comicId int64) (err error) {
	if saveOff {
		return nil
	}
	var newId int64
	result, err := linksInsertion.Exec(this.sqlId, comicId, string(this.SourceId), this.URL, this.MarkAsRead)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId
	return nil
}

func SQLSourceLinkQuery(rows *sql.Rows) (SourceLink, error) {
	var link SourceLink
	err := rows.Scan(&link.sqlId, &link.SourceId, &link.URL, &link.MarkAsRead)
	return link, err
}
