package core

import (
	"database/sql"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

var (
	scanlationInsertion *sql.Stmt
	scanlatorRelation   *sql.Stmt
	pageLinkInsertion   *sql.Stmt

	scanlationsQuery *sql.Stmt
	scanlatorsQuery  *sql.Stmt
	pageLinksQuery   *sql.Stmt
)

func init() {
	qdb.PrepareStmt(&scanlationInsertion, `
		INSERT OR REPLACE INTO scanlations(id, chapterId, sourceId, version, color, title, lang, url)
		VALUES((SELECT id FROM scanlations WHERE id = ?), ?, ?, ?, ?, ?, ?, ?);`)
	qdb.PrepareStmt(&scanlatorRelation, `INSERT OR IGNORE INTO rel_Scanlation_Scanlators(scanlationId, scanlatorId) VALUES(?, ?);`)
	qdb.PrepareStmt(&pageLinkInsertion, `INSERT INTO pages(scanlationId, url) VALUES(?, ?);`)

	qdb.PrepareStmt(&scanlationsQuery, `SELECT id, sourceId, version, color, title, lang, url FROM scanlations WHERE chapterId = ?;`)
	qdb.PrepareStmt(&scanlatorsQuery, `SELECT scanlatorId FROM rel_Scanlation_Scanlators WHERE scanlationId = ?;`)
	qdb.PrepareStmt(&pageLinksQuery, `SELECT url FROM pages WHERE scanlationId = ?;`)
}

func SQLChapterScanlationSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS scanlations(
		id INTEGER PRIMARY KEY,
		chapterId INTEGER NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
		sourceId TEXT NOT NULL,
		version INTEGER NOT NULL DEFAULT 1,
		color INTEGER NOT NULL DEFAULT 0, --bool
		title TEXT NOT NULL,
		lang INTEGER NOT NULL DEFAULT 1 REFERENCES langs(id) ON DELETE SET DEFAULT,
		url TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS scanlations_chid_idx ON scanlations(chapterId);
	CREATE TABLE IF NOT EXISTS rel_Scanlation_Scanlators(
		scanlationId INTEGER NOT NULL REFERENCES scanlations(id) ON DELETE CASCADE,
		scanlatorId INTEGER NOT NULL REFERENCES scanlators(id) ON DELETE CASCADE,
		CONSTRAINT pk_SC_SC PRIMARY KEY (scanlationId, scanlatorId)
	);
	CREATE TABLE IF NOT EXISTS pages(
		id INTEGER PRIMARY KEY,
		scanlationId INTEGER NOT NULL REFERENCES scanlations(id) ON DELETE CASCADE,
		url TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS pages_sid_idx ON pages(scanlationId);`
}

type ChapterScanlation struct {
	SourceId    SourceId
	Scanlators  JointScanlatorIds
	Version     byte
	Color       bool
	Title       string
	Language    LangId
	MetadataURL string
	PageLinks   []string

	sqlId int64
}

func (this *ChapterScanlation) sqlInsert(chapterId int64) (err error) {
	if saveOff {
		return nil
	}
	var newId int64
	result, err := scanlationInsertion.Exec(
		this.sqlId, chapterId, string(this.SourceId), this.Version, this.Color, this.Title, this.Language, this.MetadataURL,
	)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId

	for _, scanlator := range this.Scanlators.Slice() {
		result, err = scanlatorRelation.Exec(this.sqlId, scanlator)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	for _, pageLink := range this.PageLinks {
		result, err = pageLinkInsertion.Exec(this.sqlId, pageLink)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	return nil
}

func SQLChapterScanlationQuery(rows *sql.Rows) (ChapterScanlation, error) {
	scanlation := ChapterScanlation{}
	err := rows.Scan(
		&scanlation.sqlId,
		&scanlation.SourceId, &scanlation.Version, &scanlation.Color, &scanlation.Title, &scanlation.Language, &scanlation.MetadataURL,
	)
	if err != nil {
		return ChapterScanlation{}, qerr.NewLocated(err)
	}

	scanlatorRows, err := scanlatorsQuery.Query(scanlation.sqlId)
	if err != nil {
		return ChapterScanlation{}, qerr.NewLocated(err)
	}
	var scanlators []ScanlatorId
	for scanlatorRows.Next() {
		var scanlator ScanlatorId
		err = scanlatorRows.Scan(&scanlator)
		if err != nil {
			return ChapterScanlation{}, qerr.NewLocated(err)
		}
		scanlators = append(scanlators, scanlator)
	}
	scanlation.Scanlators = JoinScanlators(scanlators)

	pageLinkRows, err := pageLinksQuery.Query(scanlation.sqlId)
	if err != nil {
		return ChapterScanlation{}, qerr.NewLocated(err)
	}
	for pageLinkRows.Next() {
		var pageLink string
		err = pageLinkRows.Scan(&pageLink)
		if err != nil {
			return ChapterScanlation{}, qerr.NewLocated(err)
		}
		scanlation.PageLinks = append(scanlation.PageLinks, pageLink)
	}

	return scanlation, nil
}

type scanlationsOrder struct {
	scanlations         []ChapterScanlation
	sourceOrder         map[SourceId]linkIdx
	preferredScanlators map[JointScanlatorIds]jointIdx
	preferColor         bool
}

func (this scanlationsOrder) Len() int {
	return len(this.scanlations)
}

func (this scanlationsOrder) Less(i, j int) bool {
	csi := &this.scanlations[i]
	csj := &this.scanlations[j]

	if csi.SourceId == csj.SourceId {
		if csi.Scanlators == csj.Scanlators {
			if csi.Version == csj.Version {
				if csi.Color == csj.Color {
					return i < j
				}
				return !this.preferColor && !csi.Color && csj.Color || this.preferColor && csi.Color && !csj.Color
			}
			return csi.Version < csj.Version
		}
		orderI, okI := this.preferredScanlators[csi.Scanlators]
		orderJ, okJ := this.preferredScanlators[csj.Scanlators]
		if okI && okJ {
			return orderI < orderJ
		} else if okI {
			return true
		} else if okJ {
			return false
		}
		return i < j
	}
	return this.sourceOrder[csi.SourceId] < this.sourceOrder[csj.SourceId]
}

func (this scanlationsOrder) Swap(i, j int) {
	this.scanlations[i], this.scanlations[j] = this.scanlations[j], this.scanlations[i]
}
