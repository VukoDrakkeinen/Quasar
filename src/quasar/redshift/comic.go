package redshift

import (
	"database/sql"
	"math"
	"quasar/qutils"
	"quasar/qutils/qerr"
	. "quasar/redshift/idsdict"
	"quasar/redshift/qdb"
	"time"
)

type comicType int
type comicStatus int

const (
	InvalidComic comicType = iota
	Manga
	Manhwa
	Manhua
	Western
	Webcomic
	Other
)
const (
	ComicStatusInvalid comicStatus = iota
	ComicComplete
	ComicOngoing
	ComicOnHiatus
	ComicDiscontinued
)

const ( //SQL Statements Group keys
	comicInsertion    = "comicInsertion"
	altTitleInsertion = "altTitleInsertion"
	altTitleRelation  = "altTitleRelation"
	authorRelation    = "authorRelation"
	artistRelation    = "artistRelation"
	genreRelation     = "genreRelation"
	tagRelation       = "tagRelation"
	sourceInsertion   = "sourceInsertion"
	sourceRelation    = "sourceRelation"

	comicsQuery    = "comicsQuery"
	altTitlesQuery = "altTitlesQuery"
	authorsQuery   = "authorsQuery"
	artistsQuery   = "artistsQuery"
	genresQuery    = "genresQuery"
	tagsQuery      = "tagsQuery"
	sourcesQuery   = "sourcesQuery"
)

type sourceIndex int
type priorityIndex int
type Comic struct {
	Info     ComicInfo
	Settings IndividualSettings

	sourceIdxByPlugin map[FetcherPluginName]sourceIndex //also pluginSet
	sources           []UpdateSource                    //also pluginPriority
	chaptersOrder     ChapterIdentitiesSlice
	chapters          map[ChapterIdentity]Chapter
	scanlatorPriority []JointScanlatorIds

	sqlId int64
}

func NewComic() *Comic { //TODO: set settings
	return &Comic{
		sourceIdxByPlugin: make(map[FetcherPluginName]sourceIndex),
		chapters:          make(map[ChapterIdentity]Chapter),
	}
}

func (this *Comic) AddSource(source UpdateSource) (alreadyAdded bool) {
	return this.AddSourceAt(len(this.sources), source)
}

func (this *Comic) AddSourceAt(index int, source UpdateSource) (alreadyAdded bool) {
	existingIndex, alreadyAdded := this.sourceIdxByPlugin[source.PluginName]
	if alreadyAdded {
		source.sqlId = this.sources[existingIndex].sqlId //copy sqlId, so SQLInsert will treat new struct as old modified
		this.sources[existingIndex] = source             //replace
	} else {
		if index < len(this.sources) { //insert
			this.sources = append(this.sources, UpdateSource{}) //grow the slice
			copy(this.sources[index+1:], this.sources[index:])  //move the data we want to after our value by one
			this.sources[index] = source
		} else { //append
			this.sources = append(this.sources, source)
		}
		this.sourceIdxByPlugin[source.PluginName] = sourceIndex(index)
	}
	return
}

func (this *Comic) RemoveSource(source UpdateSource) (success bool) {
	index, exists := this.sourceIdxByPlugin[source.PluginName]
	if exists {
		this.sources = append(this.sources[:index], this.sources[index+1:]...)
	}
	return exists
}

func (this *Comic) Sources() []UpdateSource {
	ret := make([]UpdateSource, len(this.sources))
	copy(ret, this.sources)
	return ret
}

func (this *Comic) GetSource(pluginName FetcherPluginName) UpdateSource { //TODO: not found -> error?
	index := this.sourceIdxByPlugin[pluginName]
	return this.sources[index]
}

func (this *Comic) AddChapter(identity ChapterIdentity, chapter *Chapter) (merged bool) {
	this.scanlatorPriority = qutils.SetAppendSlice(this.scanlatorPriority, chapter.Scanlators()).([]JointScanlatorIds) //FIXME: purge this hack
	existingChapter, merged := this.chapters[identity]
	if merged {
		existingChapter.MergeWith(chapter)
		this.chapters[identity] = existingChapter //reinsert //TODO?: use pointers instead?
	} else {
		chapter.SetParent(this)
		this.chapters[identity] = *chapter
		this.chaptersOrder = this.chaptersOrder.Insert(this.chaptersOrder.vestedIndexOf(identity), identity)
	}
	return
}

func (this *Comic) AddMultipleChapters(identities []ChapterIdentity, chapters []Chapter) {
	minLen := int(math.Min(float64(len(identities)), float64(len(chapters))))
	nonexistentSlices := make([][]ChapterIdentity, 0, minLen/2) //Slice of slices of non-existent identities
	startIndex := 0                                             //Starting index of new slice of non-existent identities
	newStart := false                                           //Status of creation of the slice
	for i := 0; i < minLen; i++ {
		identity := identities[i]
		chapter := chapters[i]
		existingChapter, exists := this.chapters[identity]
		this.scanlatorPriority = qutils.SetAppendSlice(this.scanlatorPriority, chapter.Scanlators()).([]JointScanlatorIds) //FIXME: purge this hack
		if exists {
			existingChapter.MergeWith(&chapter)
			if newStart { //Sequence ended, add newly created slice to the list, set creation status to false
				nonexistentSlices = append(nonexistentSlices, identities[startIndex:i])
				newStart = false
			}
			this.chapters[identity] = existingChapter //reinsert //TODO?: use pointers instead?
		} else {
			chapter.SetParent(this)
			this.chapters[identity] = chapter
			if !newStart { //Sequence started, set starting index, set creation status to true
				startIndex = i
				newStart = true
			}
		}
	}
	if newStart { //Sequence ended
		nonexistentSlices = append(nonexistentSlices, identities[startIndex:])
		newStart = false
	}

	for i := 0; i < len(nonexistentSlices); i++ {
		neSlice := nonexistentSlices[i]
		insertionIndex := int(this.chaptersOrder.vestedIndexOf(neSlice[0]))
		this.chaptersOrder = this.chaptersOrder.InsertMultiple(insertionIndex, neSlice)
	}
}

func (this *Comic) GetChapter(index int) (Chapter, ChapterIdentity) { //FIXME: bounds check?
	identity := this.chaptersOrder[index]
	return this.chapters[identity], identity
}

func (this *Comic) ScanlatorsPriority() []JointScanlatorIds {
	ret := make([]JointScanlatorIds, len(this.sources))
	copy(ret, this.scanlatorPriority)
	return ret
}

func (this *Comic) SetScanlatorsPriority(priority []JointScanlatorIds) {
	this.scanlatorPriority = priority
}

func (this *Comic) ChapterCount() int {
	return len(this.chaptersOrder)
}

func (this *Comic) SQLInsert(stmts qdb.StmtGroup) (err error) {
	var newId int64
	result, err := stmts[comicInsertion].Exec(
		this.sqlId,
		this.Info.Title, this.Info.Type, this.Info.Status, this.Info.ScanlationStatus, this.Info.Description,
		this.Info.Rating, this.Info.Mature, this.Info.ThumbnailFilename,
		qutils.BoolsToBitfield(this.Settings.UseDefaults), this.Settings.UpdateNotificationMode,
		this.Settings.AccumulativeModeCount, this.Settings.DelayedModeDuration,
	)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId

	if this.Info.altSQLIds == nil {
		this.Info.altSQLIds = make(map[string]int64)
	}
	for title := range this.Info.AltTitles {
		var newATId int64
		result, err = stmts[altTitleInsertion].Exec(this.Info.altSQLIds[title], title)
		if err != nil {
			return qerr.NewLocated(err)
		}
		newATId, err = result.LastInsertId()
		if err != nil {
			return qerr.NewLocated(err)
		}
		this.Info.altSQLIds[title] = newATId
		stmts[altTitleRelation].Exec(this.sqlId, newATId)
	}

	for _, author := range this.Info.Authors {
		stmts[authorRelation].Exec(this.sqlId, author)
	}
	for _, artist := range this.Info.Artists {
		stmts[artistRelation].Exec(this.sqlId, artist)
	}
	for genre := range this.Info.Genres {
		stmts[genreRelation].Exec(this.sqlId, genre)
	}
	for tag := range this.Info.Categories {
		stmts[tagRelation].Exec(this.sqlId, tag)
	}

	for i := range this.sources {
		err = this.sources[i].SQLInsert(this.sqlId, stmts)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	for _, identity := range this.chaptersOrder {
		chapter := this.chapters[identity] //can't take a pointer
		err = chapter.SQLInsert(identity, stmts)
		if err != nil {
			return qerr.NewLocated(err)
		}
		this.chapters[identity] = chapter //so reinsert
	}

	return nil
}

func SQLComicQuery(rows *sql.Rows, stmts qdb.StmtGroup) (*Comic, error) {
	comic := NewComic()
	info := &comic.Info
	info.altSQLIds = make(map[string]int64)
	stts := IndividualSettings{}
	var comicId int64
	var thumbnailFilename sql.NullString
	var useDefaultsBitfield uint64
	var duration int64
	err := rows.Scan(
		&comic.sqlId,
		&info.Title, &info.Type, &info.Status, &info.ScanlationStatus, &info.Description, &info.Rating, &info.Mature, &thumbnailFilename,
		&useDefaultsBitfield, &stts.UpdateNotificationMode, &stts.AccumulativeModeCount, &duration,
	)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	info.ThumbnailFilename = thumbnailFilename.String
	stts.DelayedModeDuration = time.Duration(duration)
	stts.UseDefaults = qutils.BitfieldToBools(useDefaultsBitfield)
	comic.Settings = stts

	altTitleRows, err := stmts[altTitlesQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	info.AltTitles = make(map[string]struct{})
	for altTitleRows.Next() {
		var titleId int64
		var title string
		err = altTitleRows.Scan(&titleId, &title)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		info.AltTitles[title] = struct{}{}
		info.altSQLIds[title] = titleId
	}

	authorRows, err := stmts[authorsQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	for authorRows.Next() {
		var author AuthorId
		err = authorRows.Scan(&author)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		info.Authors = append(info.Authors, author)
	}

	artistRows, err := stmts[artistsQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	for artistRows.Next() {
		var artist ArtistId
		err = artistRows.Scan(&artist)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		info.Artists = append(info.Artists, artist)
	}

	genreRows, err := stmts[genresQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	info.Genres = make(map[ComicGenreId]struct{})
	for genreRows.Next() {
		var genre ComicGenreId
		err = genreRows.Scan(&genre)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		info.Genres[genre] = struct{}{}
	}

	tagRows, err := stmts[tagsQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	info.Categories = make(map[ComicTagId]struct{})
	for tagRows.Next() {
		var tag ComicTagId
		err = tagRows.Scan(&tag)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		info.Categories[tag] = struct{}{}
	}

	sourceRows, err := stmts[sourcesQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	for sourceRows.Next() {
		source, err := SQLUpdateSourceQuery(sourceRows)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		comic.AddSource(source)
	}

	var identities []ChapterIdentity
	var chapters []Chapter
	chapterRows, err := stmts[chaptersQuery].Query(comicId)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	for chapterRows.Next() {
		chapter, identity, err := SQLChapterQuery(chapterRows, stmts)
		if err != nil {
			return nil, qerr.NewLocated(err)
		}
		identities = append(identities, identity)
		chapters = append(chapters, *chapter)
	}
	comic.AddMultipleChapters(identities, chapters)

	return comic, nil
}

func SQLComicSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS comics(
		id INTEGER PRIMARY KEY,
		-- info
		title TEXT NOT NULL,
		type INTEGER NOT NULL,
		status INTEGER NOT NULL,
		scanStatus INTEGER NOT NULL,
		desc TEXT NOT NULL,
		rating REAL NOT NULL,
		mature INTEGER NOT NULL,
		thumbnailFilename TEXT,
		-- settings
		useDefaultsBits INTEGER NOT NULL,
		notifMode INTEGER,
		accumCount INTEGER,
		delayDuration INTEGER,
		downloadsPath TEXT
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_AltTitles(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		titleId INTEGER NOT NULL REFERENCES altTitles(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AT PRIMARY KEY (comicId, titleId)
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Authors(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		authorId INTEGER NOT NULL REFERENCES authors(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AU PRIMARY KEY (comicId, authorId)
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Artists(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		artistId INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AR PRIMARY KEY (comicId, artistId)
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Genres(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		genreId INTEGER NOT NULL REFERENCES genres(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_GE PRIMARY KEY (comicId, genreId)
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Tags(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		tagId INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AT PRIMARY KEY (comicId, tagId)
	);
	` + SQLUpdateSourceSchema() + SQLChapterSchema()
}

func SQLComicInsertStmts(db *qdb.QDB) (stmts qdb.StmtGroup) {
	stmts = make(qdb.StmtGroup)
	stmts[comicInsertion] = db.MustPrepare(`
		INSERT OR REPLACE INTO comics(
			id,
			title, type, status, scanStatus, desc, rating, mature, thumbnailFilename,
			useDefaultsBits, notifMode, accumCount, delayDuration
		) VALUES((SELECT id FROM comics WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	stmts[altTitleInsertion] = db.MustPrepare(`INSERT OR REPLACE INTO altTitles(id, title) VALUES((SELECT id FROM altTitles WHERE id = ?), ?);`)
	stmts[altTitleRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_AltTitles(comicId, titleId) VALUES(?, ?);`)
	stmts[authorRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Authors(comicId, authorId) VALUES(?, ?);`)
	stmts[artistRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Artists(comicId, artistId) VALUES(?, ?);`)
	stmts[genreRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Genres(comicId, genreId) VALUES(?, ?);`)
	stmts[tagRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Tags(comicId, tagId) VALUES(?, ?);`)
	sqlAddUpdateSourceInsertStmts(db, stmts)
	sqlAddChapterInsertStmts(db, stmts)
	return
}

func SQLComicQueryStmts(db *qdb.QDB) (stmts qdb.StmtGroup) {
	stmts = make(qdb.StmtGroup)
	stmts[comicsQuery] = db.MustPrepare(`
		SELECT
			id,
			title, type, status, scanStatus, desc, rating, mature, thumbnailFilename,
			useDefaultsBits, notifMode, accumCount, delayDuration
		FROM comics;`)
	stmts[altTitlesQuery] = db.MustPrepare(`SELECT id, title FROM altTitles WHERE id IN (SELECT titleId FROM rel_Comic_AltTitles WHERE comicId = ?);`)
	stmts[authorsQuery] = db.MustPrepare(`SELECT authorId FROM rel_Comic_Authors WHERE comicId = ?;`)
	stmts[artistsQuery] = db.MustPrepare(`SELECT artistId FROM rel_Comic_Artists WHERE comicId = ?;`)
	stmts[genresQuery] = db.MustPrepare(`SELECT genreId FROM rel_Comic_Genres WHERE comicId = ?;`)
	stmts[tagsQuery] = db.MustPrepare(`SELECT tagId FROM rel_Comic_Tags WHERE comicId = ?;`)
	sqlAddUpdateSourceQueryStmts(db, stmts)
	sqlAddChapterQueryStmts(db, stmts)
	return
}

type UpdateSource struct {
	PluginName FetcherPluginName
	URL        string
	MarkAsRead bool

	sqlId int64
}

func (this *UpdateSource) SQLInsert(comicId int64, stmts qdb.StmtGroup) (err error) {
	var newId int64
	result, err := stmts[sourceInsertion].Exec(this.sqlId, string(this.PluginName), this.URL, this.MarkAsRead)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.sqlId = newId
	stmts[sourceRelation].Exec(comicId, this.sqlId)
	return nil
}

func SQLUpdateSourceQuery(rows *sql.Rows) (UpdateSource, error) {
	var source UpdateSource
	err := rows.Scan(&source.sqlId, &source.PluginName, &source.URL, &source.MarkAsRead)
	return source, err
}

func SQLUpdateSourceSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS sources(
		id INTEGER PRIMARY KEY,
		pluginName TEXT NOT NULL,
		url TEXT NOT NULL,
		markAsRead INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Sources(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		sourceId INTEGER NOT NULL REFERENCES sources(id) ON DELETE CASCADE,
		CONSTRAINT pk_CO_SO PRIMARY KEY (comicId, sourceId)
	);`
}

func sqlAddUpdateSourceInsertStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[sourceInsertion] = db.MustPrepare(`
		INSERT OR REPLACE INTO sources(id, pluginName, url, markAsRead)
		VALUES((SELECT id FROM sources WHERE id = ?), ?, ?, ?);`)
	stmts[sourceRelation] = db.MustPrepare(`INSERT OR IGNORE INTO rel_Comic_Sources(comicId, sourceId) VALUES(?, ?);`)
}

func sqlAddUpdateSourceQueryStmts(db *qdb.QDB, stmts qdb.StmtGroup) {
	stmts[sourcesQuery] = db.MustPrepare(`
	SELECT id, pluginName, url, markAsRead
	FROM sources
	WHERE id IN(SELECT sourceId FROM rel_Comic_Sources WHERE comicId = ?);`)
}

type ComicInfo struct {
	Title             string
	AltTitles         map[string]struct{}
	Authors           []AuthorId
	Artists           []ArtistId
	Genres            map[ComicGenreId]struct{}
	Categories        map[ComicTagId]struct{}
	Type              comicType
	Status            comicStatus
	ScanlationStatus  ScanlationStatus
	Description       string
	Rating            float32
	Mature            bool
	ThumbnailFilename string

	altSQLIds map[string]int64
}

func (this *ComicInfo) MergeWith(another *ComicInfo) {
	if this.AltTitles == nil {
		this.AltTitles = make(map[string]struct{})
		this.Genres = make(map[ComicGenreId]struct{})
		this.Categories = make(map[ComicTagId]struct{})
	}

	if this.Title == "" {
		this.Title = another.Title
	}

	for altTitle, _ := range another.AltTitles {
		this.AltTitles[altTitle] = struct{}{}
	}

	authorsSet := make(map[AuthorId]struct{})
	for _, author := range this.Authors {
		authorsSet[author] = struct{}{}
	}
	for _, author := range another.Authors {
		if _, exists := authorsSet[author]; !exists {
			this.Authors = append(this.Authors, author)
		}
	}

	artistsSet := make(map[ArtistId]struct{})
	for _, artist := range this.Artists {
		artistsSet[artist] = struct{}{}
	}
	for _, artist := range another.Artists {
		if _, exists := artistsSet[artist]; !exists {
			this.Artists = append(this.Artists, artist)
		}
	}

	for genre, _ := range another.Genres {
		this.Genres[genre] = struct{}{}
	}

	for tag, _ := range another.Categories {
		this.Categories[tag] = struct{}{}
	}

	if this.Type == InvalidComic {
		this.Type = another.Type
	}

	if (this.Status == ComicStatusInvalid) ||
		(this.Status == ComicOngoing && another.Status == ComicOnHiatus) ||
		(this.Status == ComicOnHiatus && another.Status == ComicOngoing) ||
		(another.Status == ComicDiscontinued) ||
		(another.Status == ComicComplete) {
		this.Status = another.Status
	}

	if (this.ScanlationStatus == ScanlationStatusInvalid) ||
		(this.ScanlationStatus == ScanlationOngoing && another.ScanlationStatus == ScanlationOnHiatus) ||
		(this.ScanlationStatus == ScanlationOnHiatus && another.ScanlationStatus == ScanlationOngoing) ||
		(another.ScanlationStatus == ScanlationDropped) ||
		(another.ScanlationStatus == ScanlationInDesperateNeedOfMoreStaff) ||
		(another.ScanlationStatus == ScanlationComplete) {
		this.ScanlationStatus = another.ScanlationStatus
	}

	if this.Description == "" {
		this.Description = another.Description
	}

	if this.Rating == 0 {
		this.Rating = another.Rating
	} else if another.Rating != 0 {
		this.Rating = (this.Rating + another.Rating) / 2
	}

	this.Mature = another.Mature || this.Mature

	if this.ThumbnailFilename == "" {
		this.ThumbnailFilename = another.ThumbnailFilename
	}
}
