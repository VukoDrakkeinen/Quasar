package core

import (
	"database/sql"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

const ( //SQL Statements Group keys
	chapterInsertion    = "chapterInsertion"
	chapterRelation     = "chapterRelation"
	scanlationInsertion = "scanlationInsertion"
	scanlationRelation  = "scanlationRelation"
	scanlatorRelation   = "scanlatorRelation"
	pageLinkInsertion   = "pageLinkInsertion"
	pageLinkRelation    = "pageLinkRelation"

	chaptersQuery    = "chaptersQuery"
	scanlationsQuery = "scanlationsQuery"
	scanlatorsQuery  = "scanlatorsQuery"
	pageLinksQuery   = "pageLinksQuery"
)

const (
	LQ_Modifier byte = 10 * iota
	MQ_Modifier
	HQ_Modifier
)

type scanlationIndex int
type Chapter struct {
	parent      *Comic
	scanlations []ChapterScanlation
	mapping     map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex
	usedPlugins []FetcherPluginName
	AlreadyRead bool

	sqlId int64
}

func NewChapter(alreadyRead bool) *Chapter {
	return &Chapter{
		AlreadyRead: alreadyRead,
		mapping:     make(map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex),
	}

}

func (this *Chapter) Scanlation(index int) ChapterScanlation {
	pluginName, scanlators := this.indexToPath(index)
	return this.scanlations[this.mapping[pluginName][scanlators]]
}

func (this *Chapter) ScanlationsCount() int {
	return len(this.scanlations)
}

func (this *Chapter) MergeWith(another *Chapter) *Chapter {
	this.AlreadyRead = another.AlreadyRead || this.AlreadyRead
	for _, scanlation := range another.scanlations {
		this.AddScanlation(scanlation)
	}
	return this
}

func (this *Chapter) AddScanlation(scanlation ChapterScanlation) (replaced bool) {
	if mapped, pluginExists := this.mapping[scanlation.PluginName]; pluginExists {
		if index, jointExists := mapped[scanlation.Scanlators]; jointExists {
			scanlation.sqlId = this.scanlations[index].sqlId //copy sqlId, so SQLInsert will treat new struct as old modified
			this.scanlations[index] = scanlation             //replace
			return true
		}
	} else {
		this.usedPlugins = append(this.usedPlugins, scanlation.PluginName)
	}

	if this.mapping[scanlation.PluginName] == nil { //TODO: refactor
		this.mapping[scanlation.PluginName] = make(map[JointScanlatorIds]scanlationIndex)
	}
	this.mapping[scanlation.PluginName][scanlation.Scanlators] = scanlationIndex(len(this.scanlations))
	this.scanlations = append(this.scanlations, scanlation)
	return false
}

func (this *Chapter) RemoveScanlation(index int) {
	pluginName, scanlators := this.indexToPath(index)
	realIndex := this.mapping[pluginName][scanlators]

	this.scanlations = append(this.scanlations[:realIndex], this.scanlations[realIndex+1:]...)
	delete(this.mapping[pluginName], scanlators)

	if len(this.mapping[pluginName]) == 0 {
		delete(this.mapping, pluginName)
		deletionIndex, _ := qutils.IndexOf(this.usedPlugins, pluginName)
		this.usedPlugins = append(this.usedPlugins[:deletionIndex], this.usedPlugins[deletionIndex+1:]...)
	}
}

func (this *Chapter) RemoveScanlationsForPlugin(pluginName FetcherPluginName) {
	for _, realIndex := range this.mapping[pluginName] {
		this.scanlations = append(this.scanlations[:realIndex], this.scanlations[realIndex+1:]...)
	}
	delete(this.mapping, pluginName)
	deletionIndex, _ := qutils.IndexOf(this.usedPlugins, pluginName)
	this.usedPlugins = append(this.usedPlugins[:deletionIndex], this.usedPlugins[deletionIndex+1:]...)
}

func (this *Chapter) Scanlators() (ret []JointScanlatorIds) {
	if this.parent != nil {
		for _, pluginName := range this.usedPlugins {
			perPlugin := this.mapping[pluginName]
			for _, scanlator := range this.parent.ScanlatorsPriority() {
				if _, exists := perPlugin[scanlator]; exists {
					ret = append(ret, scanlator)
				}
			}
		}
	} else {
		for _, scanlation := range this.scanlations {
			ret = append(ret, scanlation.Scanlators)
		}
	}
	return
}

func (this *Chapter) SetParent(comic *Comic) {
	this.parent = comic
}

func (this *Chapter) indexToPath(index int) (FetcherPluginName, JointScanlatorIds) {
	if this.parent == nil { //We have no parent, so we can't access priority lists for plugins and scanlators
		scanlation := this.scanlations[index]
		return scanlation.PluginName, scanlation.Scanlators
	}

	var pluginNames []FetcherPluginName            //Create a set of plugin names with prioritized ones at the beginning
	for _, source := range this.parent.Sources() { //Add prioritized plugin names
		if _, exists := this.mapping[source.PluginName]; exists {
			pluginNames = append(pluginNames, source.PluginName)
		}
	}
	for _, pluginName := range this.usedPlugins { //Add the rest
		if !this.parent.UsesPlugin(pluginName) {
			pluginNames = append(pluginNames, pluginName)
		}
	}

	var pluginName FetcherPluginName
	for _, pluginName = range pluginNames { //Absolute index => relative index
		jointsPerPlugin := len(this.mapping[pluginName])
		if index >= jointsPerPlugin {
			index -= jointsPerPlugin
		} else {
			break
		}
	}

	scanlatorSet := this.mapping[pluginName]
	var scanlators []JointScanlatorIds
	for _, scanlator := range this.parent.ScanlatorsPriority() { //Create a set of this chapter's scanlators (prioritized first)
		if _, exists := scanlatorSet[scanlator]; exists {
			scanlators = append(scanlators, scanlator)
		}
	}

	return pluginName, scanlators[index]
}

func (this *Chapter) SQLInsert(identity ChapterIdentity, stmts qdb.StmtGroup) (err error) {
	var newId int64
	result, err := stmts[chapterInsertion].Exec(this.sqlId, identity.n(), this.AlreadyRead)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId

	result, err = stmts[chapterRelation].Exec(this.parent.SQLId(), this.sqlId)
	if err != nil {
		return qerr.NewLocated(err)
	}

	for i := range this.scanlations {
		err = this.scanlations[i].SQLInsert(this.sqlId, stmts)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	return nil
}

func SQLChapterQuery(rows *sql.Rows, stmts qdb.StmtGroup) (*Chapter, ChapterIdentity, error) {
	var identity ChapterIdentity
	chapter := NewChapter(false)
	err := rows.Scan(&chapter.sqlId, &identity, &chapter.AlreadyRead)
	if err != nil {
		return nil, identity, qerr.NewLocated(err)
	}

	scanlationRows, err := stmts[scanlationsQuery].Query(chapter.sqlId)
	if err != nil {
		return nil, identity, qerr.NewLocated(err)
	}
	for scanlationRows.Next() {
		scanlation, err := SQLChapterScanlationQuery(scanlationRows, stmts)
		if err != nil {
			return nil, identity, qerr.NewLocated(err)
		}
		chapter.AddScanlation(*scanlation)
	}

	return chapter, identity, err
}

func SQLChapterSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS chapters(
		id INTEGER PRIMARY KEY,
		identity INTEGER NOT NULL,
		alreadyRead INTEGER NOT NULL,
		CONSTRAINT uq_CH_ID UNIQUE (id, identity)
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Chapters(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		chapterId INTEGER NOT NULL UNIQUE REFERENCES chapters(id) ON DELETE CASCADE,
		CONSTRAINT pk_CO_CH PRIMARY KEY (comicId, chapterId)
	);
	` + SQLChapterScanlationSchema()
}

func sqlAddChapterInsertStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[chapterInsertion] = db.MustPrepare(`
		INSERT OR REPLACE INTO chapters(id, identity, alreadyRead)
		VALUES((SELECT id FROM chapters WHERE id = ?), ?, ?);`)
	stmts[chapterRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Chapters(comicId, chapterId) VALUES(?, ?);`)
	sqlAddChapterScanlationInsertStmts(db, stmts)
}

func sqlAddChapterQueryStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[chaptersQuery] = db.MustPrepare(`
		SELECT id, identity, alreadyRead
		FROM chapters
		WHERE id IN(SELECT chapterId FROM rel_Comic_Chapters WHERE comicId = ?);`)
	sqlAddChapterScanlationQueryStmts(db, stmts)
}

type ChapterScanlation struct {
	Title      string
	Language   LangId
	Scanlators JointScanlatorIds //TODO: see scanlators.go
	PluginName FetcherPluginName
	URL        string
	PageLinks  []string

	plSQLIds []int64
	sqlId    int64
}

func (this *ChapterScanlation) SQLInsert(chapterId int64, stmts qdb.StmtGroup) (err error) {
	var newId int64
	result, err := stmts[scanlationInsertion].Exec(this.sqlId, this.Title, this.Language, string(this.PluginName), this.URL)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId

	result, err = stmts[scanlationRelation].Exec(chapterId, this.sqlId)
	if err != nil {
		return qerr.NewLocated(err)
	}

	for _, scanlator := range this.Scanlators.ToSlice() {
		result, err = stmts[scanlatorRelation].Exec(this.sqlId, scanlator)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	if this.plSQLIds == nil {
		this.plSQLIds = make([]int64, len(this.PageLinks))
	}
	for i, pageLink := range this.PageLinks {
		var pageLinkId int64 = this.plSQLIds[i] //WARNING: may go out of bounds (shouldn't ever; leaving it for the sake of experiment)
		result, err = stmts[pageLinkInsertion].Exec(pageLink)
		if err != nil {
			return qerr.NewLocated(err)
		}
		pageLinkId, err = result.LastInsertId()
		if err != nil {
			return qerr.NewLocated(err)
		}
		this.plSQLIds[i] = pageLinkId
		stmts[pageLinkRelation].Exec(this.sqlId, pageLinkId)
	}

	return nil
}

func SQLChapterScanlationQuery(rows *sql.Rows, stmts qdb.StmtGroup) (*ChapterScanlation, error) {
	scanlation := &ChapterScanlation{}
	err := rows.Scan(&scanlation.sqlId, &scanlation.Title, &scanlation.Language, &scanlation.PluginName, &scanlation.URL)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}

	scanlatorRows, err := stmts[scanlatorsQuery].Query(scanlation.sqlId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	var scanlators []ScanlatorId
	for scanlatorRows.Next() {
		var scanlator ScanlatorId
		err = scanlatorRows.Scan(&scanlator)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		scanlators = append(scanlators, scanlator)
	}
	scanlation.Scanlators = JoinScanlators(scanlators)

	pageLinkRows, err := stmts[pageLinksQuery].Query(scanlation.sqlId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	for pageLinkRows.Next() {
		var pageLinkId int64
		var pageLink string
		err = pageLinkRows.Scan(&pageLinkId, &pageLink)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		scanlation.PageLinks = append(scanlation.PageLinks, pageLink)
		scanlation.plSQLIds = append(scanlation.plSQLIds, pageLinkId)
	}

	return scanlation, nil
}

func SQLChapterScanlationSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS scanlations(
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		lang INTEGER NOT NULL DEFAULT 1 REFERENCES langs(id) ON DELETE SET DEFAULT,
		pluginName TEXT NOT NULL,
		url TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Chapter_Scanlations(
		chapterId INTEGER NOT NULL REFERENCES chapters(id) ON DELETE CASCADE,
		scanlationId INTEGER NOT NULL REFERENCES scanlations(id) ON DELETE CASCADE,
		CONSTRAINT pk_CH_SC PRIMARY KEY (chapterId, scanlationId)
	);
	CREATE TABLE IF NOT EXISTS rel_Scanlation_Scanlators(
		scanlationId INTEGER NOT NULL REFERENCES scanlations(id) ON DELETE CASCADE,
		scanlatorId INTEGER NOT NULL REFERENCES scanlators(id) ON DELETE CASCADE,
		CONSTRAINT pk_SC_SC PRIMARY KEY (scanlationId, scanlatorId)
	);
	CREATE TABLE IF NOT EXISTS pageLinks(
		id INTEGER PRIMARY KEY,
		url TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Scanlation_PageLinks(
		scanlationId INTEGER NOT NULL REFERENCES scanlations(id) ON DELETE CASCADE,
		pageLinkId INTEGER NOT NULL REFERENCES pageLinks(id) ON DELETE CASCADE,
		CONSTRAINT pk_SC_PA PRIMARY KEY (scanlationId, pageLinkId)
	);`
}

func sqlAddChapterScanlationInsertStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[scanlationInsertion] = db.MustPrepare(`
		INSERT OR REPLACE INTO scanlations(id, title, lang, pluginName, url)
		VALUES((SELECT id FROM scanlations WHERE id = ?), ?, ?, ?, ?);`)
	stmts[scanlationRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Chapter_Scanlations(chapterId, scanlationid) VALUES(?, ?);`)
	stmts[scanlatorRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Scanlation_Scanlators(scanlationId, scanlatorId) VALUES(?, ?);`)
	stmts[pageLinkInsertion] = db.MustPrepare(`INSERT OR REPLACE INTO pageLinks(id, url) VALUES((SELECT id FROM pageLinks WHERE id = ?), ?);`)
	stmts[pageLinkRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Scanlation_PageLinks(scanlationId, pageLinkId) VALUES(?, ?);`)
}

func sqlAddChapterScanlationQueryStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[scanlationsQuery] = db.MustPrepare(`
		SELECT id, title, lang, pluginName, url
		FROM scanlations
		WHERE id IN(SELECT scanlationId FROM rel_Chapter_Scanlations WHERE chapterId = ?);`)
	stmts[scanlatorsQuery] = db.MustPrepare(`SELECT scanlatorId FROM rel_Scanlation_Scanlators WHERE scanlationId = ?;`)
	stmts[pageLinksQuery] = db.MustPrepare(`
		SELECT id, url
		FROM pageLinks
		WHERE id IN(SELECT pageLinkId FROM rel_Scanlation_PageLinks WHERE scanlationId = ?);`)
}
