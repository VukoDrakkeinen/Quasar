package core

import (
	"database/sql"
	"sort"

	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

var (
	chapterInsertion        *sql.Stmt
	chaptersQuery           *sql.Stmt
	chapterChangeReadStatus *sql.Stmt
)

func init() {
	qdb.PrepareStmt(&chapterInsertion, `
			INSERT OR REPLACE INTO chapters(id, comicId, identity, alreadyRead)
			VALUES((SELECT id FROM chapters WHERE id = ?), ?, ?, ?);`) //todo: remove optional id?
	qdb.PrepareStmt(&chaptersQuery, `SELECT id, identity, alreadyRead FROM chapters WHERE comicId = ?;`)
	qdb.PrepareStmt(&chapterChangeReadStatus, `UPDATE chapters SET alreadyRead = ?2 WHERE id = ?1;`)
}

func SQLChapterSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS chapters(
		id INTEGER PRIMARY KEY,
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		identity INTEGER NOT NULL,
		alreadyRead INTEGER NOT NULL --bool
	);
	CREATE INDEX IF NOT EXISTS chapters_cid_idx ON chapters(comicId);
	` + SQLChapterScanlationSchema()
	//--CONSTRAINT uq_CH_ID UNIQUE (id, identity)
}

type Chapter struct {
	MarkedRead bool

	parent        *Comic
	scanlations   []ChapterScanlation
	orderCacheKey cacheKey

	sqlId int64
}
type IdentifiedChapter struct {
	chapter  Chapter
	identity ChapterIdentity
}

func (this *Chapter) reorderScanlations() {
	sort.Sort(scanlationsOrder{
		this.scanlations, this.parent.PreferredSources(), this.parent.PreferredScanlators(), this.parent.PrefersColor(),
	})
	this.orderCacheKey = this.parent.preferencesCacheKey()
}

func (this *Chapter) Scanlation(index int) ChapterScanlation {
	if this.parent != nil && this.parent.preferencesStale(this.orderCacheKey) {
		this.reorderScanlations()
	}
	return this.scanlations[index]
}

func (this *Chapter) ScanlationsCount() int {
	return len(this.scanlations)
}

func (this *Chapter) MergeWith(another *Chapter) *Chapter {
	if this.sqlId == 0 {
		//this.sqlInsert(parent, identity)	//todo
	}

	prevRead := this.MarkedRead
	this.MarkedRead = another.MarkedRead || this.MarkedRead
	if !saveOff && this.MarkedRead != prevRead {
		chapterChangeReadStatus.Exec(this.sqlId, this.MarkedRead)
	}

	for _, scanlation := range another.scanlations {
		replaced := this.AddScanlation(scanlation)
		if !saveOff && !replaced {
			scanlation.sqlInsert(this.sqlId)
		} else {
			//todo: handle scanlation replacement
		}
	}
	return this
}

func (this *Chapter) AddScanlation(scanlation ChapterScanlation) (replaced bool) {
	exists := false //todo: detect existing scanlations
	if exists {
		index := 0
		scanlation.sqlId = this.scanlations[index].sqlId
		this.scanlations[index] = scanlation
		return true
	}
	this.scanlations = append(this.scanlations, scanlation)
	return false
}

func (this *Chapter) RemoveScanlation(index int) {
	this.scanlations = append(this.scanlations[:index], this.scanlations[index+1:]...)
}

func (this *Chapter) RemoveScanlationsForPlugin(pluginName SourceId) {
	for index, scanlation := range this.scanlations {
		if scanlation.SourceId == pluginName {
			this.scanlations = append(this.scanlations[:index], this.scanlations[index+1:]...)
		}
	}
}

func (this *Chapter) Scanlators() (ret []JointScanlatorIds) {
	for _, scanlation := range this.scanlations {
		ret = append(ret, scanlation.Scanlators)
	}
	return
}

func (this *Chapter) setParent(comic *Comic) {
	this.parent = comic
}

func (this *Chapter) sqlInsert(parent *Comic, identity ChapterIdentity) (err error) {
	if saveOff {
		return nil
	}
	var newId int64
	result, err := chapterInsertion.Exec(this.sqlId, parent.SQLId(), identity.n(), this.MarkedRead)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId

	for i := range this.scanlations {
		err = this.scanlations[i].sqlInsert(this.sqlId)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	return nil
}

func SQLChapterQuery(rows *sql.Rows) (Chapter, ChapterIdentity, error) {
	var identity ChapterIdentity
	chapter := Chapter{MarkedRead: false}
	err := rows.Scan(&chapter.sqlId, &identity, &chapter.MarkedRead)
	if err != nil {
		return Chapter{}, ChapterIdentity{}, qerr.NewLocated(err)
	}

	scanlationRows, err := scanlationsQuery.Query(chapter.sqlId)
	if err != nil {
		return Chapter{}, ChapterIdentity{}, qerr.NewLocated(err)
	}
	for scanlationRows.Next() {
		scanlation, err := SQLChapterScanlationQuery(scanlationRows)
		if err != nil {
			return Chapter{}, ChapterIdentity{}, qerr.NewLocated(err)
		}
		chapter.AddScanlation(scanlation)
	}

	return chapter, identity, nil
}
