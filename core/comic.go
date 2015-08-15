package core

import (
	"database/sql"
	"fmt"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/math"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

type comicType int
type comicStatus int
type ScanlationStatus int

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
const (
	ScanlationStatusInvalid ScanlationStatus = iota
	ScanlationComplete
	ScanlationOngoing
	ScanlationOnHiatus
	ScanlationDropped
	ScanlationInDesperateNeedOfMoreStaff
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

type syncRWMutex struct {
	internal syncRWMutex
}

func (this *syncRWMutex) Lock() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("#+Locking at %s:%d\n", file, line)
	this.internal.Lock()
}
func (this *syncRWMutex) Unlock() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("#-Unlocking at %s:%d\n", file, line)
	this.internal.Unlock()
}
func (this *syncRWMutex) RLock() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("#+RLocking at %s:%d\n", file, line)
	this.internal.RLock()
}
func (this *syncRWMutex) RUnlock() {
	_, file, line, _ := runtime.Caller(1)
	fmt.Printf("#-RUnlocking at %s:%d\n", file, line)
	this.internal.RUnlock()
} //*/

type Comic struct { //TODO: rework all that concurrency-safeness (is that even a word), it's a mess
	info     ComicInfo
	settings IndividualSettings

	sourceIdxByPlugin map[FetcherPluginName]sourceIndex //also pluginSet
	sources           []UpdateSource                    //also pluginPriority
	chaptersOrder     ChapterIdentitiesSlice
	chapters          map[ChapterIdentity]Chapter
	scanlatorPriority []JointScanlatorIds
	cachedReadCount   int

	sqlId int64

	lock sync.RWMutex
}
type sourceIndex int
type priorityIndex int

func NewComic(settings IndividualSettings) *Comic {
	return &Comic{
		settings:          settings,
		sourceIdxByPlugin: make(map[FetcherPluginName]sourceIndex),
		chapters:          make(map[ChapterIdentity]Chapter),
		cachedReadCount:   -1,
	}
}

func (this *Comic) Info() ComicInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.info
}

func (this *Comic) SetInfo(info ComicInfo) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.info = info
}

func (this *Comic) Settings() IndividualSettings {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.settings
}

func (this *Comic) SetSettings(stts IndividualSettings) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.settings = stts
}

func (this *Comic) AddSource(source UpdateSource) (alreadyAdded bool) {
	return this.AddSourceAt(len(this.sources), source)
}

func (this *Comic) AddSourceAt(index int, source UpdateSource) (alreadyAdded bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
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
	this.lock.Lock()
	defer this.lock.Unlock()
	index, exists := this.sourceIdxByPlugin[source.PluginName]
	if exists {
		this.sources = append(this.sources[:index], this.sources[index+1:]...)
	}
	return exists
}

func (this *Comic) Sources() []UpdateSource {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ret := make([]UpdateSource, len(this.sources))
	copy(ret, this.sources)
	return ret
}

func (this *Comic) GetSource(pluginName FetcherPluginName) UpdateSource { //TODO: not found -> error?
	this.lock.RLock()
	defer this.lock.RUnlock()
	index := this.sourceIdxByPlugin[pluginName]
	return this.sources[index]
}

func (this *Comic) AddChapter(identity ChapterIdentity, chapter *Chapter) (merged bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.scanlatorPriority = qutils.SetAppendSlice(this.scanlatorPriority, chapter.Scanlators()).([]JointScanlatorIds) //TODO FIXME: purge this hack
	existingChapter, merged := this.chapters[identity]
	if merged {
		existingChapter.MergeWith(chapter)
		this.chapters[identity] = existingChapter //reinsert //TODO?: use pointers instead?
	} else {
		chapter.SetParent(this)
		this.chapters[identity] = *chapter
		this.chaptersOrder = this.chaptersOrder.Insert(this.chaptersOrder.vestedIndexOf(identity), identity)
		if chapter.AlreadyRead {
			this.cachedReadCount += 1
		}
	}
	return
}

func (this *Comic) AddMultipleChapters(identities []ChapterIdentity, chapters []Chapter) {
	this.lock.Lock()
	defer this.lock.Unlock()
	if len(identities) != len(chapters) {
		qlog.Log(qlog.Warning, "Comic.AddMultipleChapters: provided slices lengths do not match!")
	}
	minLen := int(math.Min(int64(len(identities)), int64(len(chapters))))
	nonexistentSlices := make([][]ChapterIdentity, 0, minLen/2) //Slice of slices of non-existent identities
	startIndex := 0                                             //Starting index of new slice of non-existent identities
	newStart := false                                           //Status of creation of the slice
	for i := 0; i < minLen; i++ {
		identity := identities[i]
		chapter := chapters[i]
		existingChapter, exists := this.chapters[identity]
		this.scanlatorPriority = qutils.SetAppendSlice(this.scanlatorPriority, chapter.Scanlators()).([]JointScanlatorIds) //TODO FIXME: purge this hack
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
			if chapter.AlreadyRead {
				this.cachedReadCount += 1
			}
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

	for _, neSlice := range nonexistentSlices {
		insertionIndex := int(this.chaptersOrder.vestedIndexOf(neSlice[0]))
		this.chaptersOrder = this.chaptersOrder.InsertMultiple(insertionIndex, neSlice)
	}
}

func (this *Comic) GetChapter(index int) (Chapter, ChapterIdentity) { //TODO FIXME: bounds check?
	this.lock.RLock()
	defer this.lock.RUnlock()
	identity := this.chaptersOrder[index]
	return this.chapters[identity], identity
}

func (this *Comic) ScanlatorsPriority() []JointScanlatorIds {
	this.lock.RLock()
	defer this.lock.RUnlock()
	ret := make([]JointScanlatorIds, len(this.sources))
	copy(ret, this.scanlatorPriority)
	return ret
}

func (this *Comic) SetScanlatorsPriority(priority []JointScanlatorIds) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.scanlatorPriority = priority
}

func (this *Comic) ChapterCount() int { //TODO: rename ChaptersCount()
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.chaptersOrder)
}

func (this *Comic) ChaptersReadCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if this.cachedReadCount != -1 {
		return this.cachedReadCount
	}
	var readCount int
	chapterCount := this.ChapterCount() //IT'S CALLED EVERY ITERATION?! O_O
	for i := 0; i < chapterCount; i++ {

		if chapter, _ := this.GetChapter(i); chapter.AlreadyRead {
			readCount++
		}
	}
	this.lock.RUnlock()
	this.lock.Lock()
	this.cachedReadCount = readCount
	this.lock.Unlock()
	this.lock.RLock()
	return readCount
}

func (this *Comic) SQLId() int64 {
	return atomic.LoadInt64(&this.sqlId)
}

func (this *Comic) SQLInsert(stmts qdb.StmtGroup) (err error) {
	this.lock.RLock()
	var newId int64
	result, err := stmts[comicInsertion].Exec(
		this.sqlId,
		this.info.Title, this.info.Type, this.info.Status, this.info.ScanlationStatus, this.info.Description,
		this.info.Rating, this.info.Mature, this.info.ThumbnailFilename,
		qutils.BoolsToBitfield(this.settings.OverrideDefaults), this.settings.FetchOnStartup,
		this.settings.IntervalFetching, this.settings.FetchFrequency, this.settings.NotificationMode,
		this.settings.AccumulativeModeCount, this.settings.DelayedModeDuration,
	)
	if err != nil {
		return qerr.NewLocated(err)
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return qerr.NewLocated(err)
	}
	this.lock.RUnlock()
	this.lock.Lock()
	this.sqlId = newId

	if this.info.titlesSQLIds == nil {
		this.info.titlesSQLIds = make(map[string]int64)
	}
	for title := range this.info.AltTitles {
		var newATId int64
		result, err = stmts[altTitleInsertion].Exec(this.info.titlesSQLIds[title], title)
		if err != nil {
			return qerr.NewLocated(err)
		}
		newATId, err = result.LastInsertId()
		if err != nil {
			return qerr.NewLocated(err)
		}
		this.info.titlesSQLIds[title] = newATId
		stmts[altTitleRelation].Exec(this.sqlId, newATId)
	}
	this.lock.Unlock()
	this.lock.RLock()

	for _, author := range this.info.Authors {
		stmts[authorRelation].Exec(this.sqlId, author)
	}
	for _, artist := range this.info.Artists {
		stmts[artistRelation].Exec(this.sqlId, artist)
	}
	for genre := range this.info.Genres {
		stmts[genreRelation].Exec(this.sqlId, genre)
	}
	for tag := range this.info.Categories {
		stmts[tagRelation].Exec(this.sqlId, tag)
	}

	for i := range this.sources {
		err = this.sources[i].SQLInsert(this.sqlId, stmts)
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	this.lock.RUnlock()
	this.lock.Lock()
	for _, identity := range this.chaptersOrder {
		chapter := this.chapters[identity] //can't take a pointer
		err = chapter.SQLInsert(identity, stmts)
		if err != nil {
			this.lock.Unlock()
			return qerr.NewLocated(err)
		}
		this.chapters[identity] = chapter //so reinsert
	}
	this.lock.Unlock()

	return nil
}

func SQLComicQuery(rows *sql.Rows, stmts qdb.StmtGroup) (*Comic, error) {
	comic := NewComic(IndividualSettings{})
	info := &comic.info
	info.titlesSQLIds = make(map[string]int64)
	stts := IndividualSettings{}
	var comicId int64
	var thumbnailFilename sql.NullString
	var overrideDefaultsBitfield uint64
	var fetchFreq int64
	var duration int64
	err := rows.Scan(
		&comic.sqlId,
		&info.Title, &info.Type, &info.Status, &info.ScanlationStatus, &info.Description, &info.Rating, &info.Mature, &thumbnailFilename,
		&overrideDefaultsBitfield, &stts.FetchOnStartup, &stts.IntervalFetching, &fetchFreq, &stts.NotificationMode,
		&stts.AccumulativeModeCount, &duration,
	)
	if err != nil {
		return nil, qerr.NewLocated(err)
	}
	comicId = comic.sqlId
	info.ThumbnailFilename = thumbnailFilename.String
	stts.DelayedModeDuration = time.Duration(duration)
	stts.OverrideDefaults = qutils.BitfieldToBools(overrideDefaultsBitfield, Bitlength(ComicSettings))
	stts.FetchFrequency = time.Duration(fetchFreq)
	comic.settings = stts //TODO?: merge settings, so loaded won't overwrite new defaults?

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
		info.titlesSQLIds[title] = titleId
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
		fetchOnStartup INTEGER,
		intervalFetching INTEGER,
		fetchFrequency INTEGER,
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
			useDefaultsBits, fetchOnStartup, intervalFetching, fetchFrequency, notifMode, accumCount, delayDuration
		) VALUES((SELECT id FROM comics WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
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
			useDefaultsBits, fetchOnStartup, intervalFetching, fetchFrequency, notifMode, accumCount, delayDuration
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

	titlesSQLIds map[string]int64
}

//func (this *ComicInfo) MergeWith(another *ComicInfo) (merged *ComicInfo) {
func (this ComicInfo) MergeWith(another *ComicInfo) (merged *ComicInfo) {
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

	return &this
}
