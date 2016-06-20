package core

import (
	"database/sql"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"

	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/math"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
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

var (
	comicInsertion     *sql.Stmt
	titleInsertion     *sql.Stmt
	authorRelation     *sql.Stmt
	artistRelation     *sql.Stmt
	genreRelation      *sql.Stmt
	tagRelation        *sql.Stmt
	thumbnailInsertion *sql.Stmt

	comicsQuery     *sql.Stmt
	titlesQuery     *sql.Stmt
	authorsQuery    *sql.Stmt
	artistsQuery    *sql.Stmt
	genresQuery     *sql.Stmt
	tagsQuery       *sql.Stmt
	thumbnailsQuery *sql.Stmt
)

func init() {
	qdb.PrepareStmt(&comicInsertion, `
		INSERT OR REPLACE INTO comics(
			id,
			titleIdx, type, status, scanStatus, desc, rating, mature, thumbnailIdx,
			useDefaultsBits, fetchOnStartup, intervalFetching, fetchFrequency, notifMode, accumCount, delayDuration
		) VALUES((SELECT id FROM comics WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`)
	qdb.PrepareStmt(&titleInsertion, `INSERT OR REPLACE INTO titles(comicId, title) VALUES(?, ?);`)
	qdb.PrepareStmt(&thumbnailInsertion, `
		INSERT OR REPLACE INTO thumbnails(id, comicId, thumbnail) VALUES((SELECT id FROM thumbnails WHERE id = ?), ?, ?);`)
	qdb.PrepareStmt(&authorRelation, `INSERT OR IGNORE INTO rel_Comic_Authors(comicId, authorId) VALUES(?, ?);`)
	qdb.PrepareStmt(&artistRelation, `INSERT OR IGNORE INTO rel_Comic_Artists(comicId, artistId) VALUES(?, ?);`)
	qdb.PrepareStmt(&genreRelation, `INSERT OR IGNORE INTO rel_Comic_Genres(comicId, genreId) VALUES(?, ?);`)
	qdb.PrepareStmt(&tagRelation, `INSERT OR IGNORE INTO rel_Comic_Tags(comicId, tagId) VALUES(?, ?);`)

	qdb.PrepareStmt(&comicsQuery, `
		SELECT
			id,
			titleIdx, type, status, scanStatus, desc, rating, mature, thumbnailIdx,
			useDefaultsBits, fetchOnStartup, intervalFetching, fetchFrequency, notifMode, accumCount, delayDuration
		FROM comics;`)
	qdb.PrepareStmt(&titlesQuery, `SELECT title FROM titles WHERE comicId = ?;`)
	qdb.PrepareStmt(&authorsQuery, `SELECT authorId FROM rel_Comic_Authors WHERE comicId = ?;`)
	qdb.PrepareStmt(&artistsQuery, `SELECT artistId FROM rel_Comic_Artists WHERE comicId = ?;`)
	qdb.PrepareStmt(&genresQuery, `SELECT genreId FROM rel_Comic_Genres WHERE comicId = ?;`)
	qdb.PrepareStmt(&tagsQuery, `SELECT tagId FROM rel_Comic_Tags WHERE comicId = ?;`)
	qdb.PrepareStmt(&thumbnailsQuery, `SELECT thumbnail FROM thumbnails WHERE comicId = ?;`)
}

func SQLComicSchema() string {
	return `
	CREATE TABLE IF NOT EXISTS comics(
		id INTEGER PRIMARY KEY,
		-- info
		titleIdx INTEGER NOT NULL,
		type INTEGER NOT NULL,
		status INTEGER NOT NULL,
		scanStatus INTEGER NOT NULL,
		desc TEXT NOT NULL,
		rating INTEGER NOT NULL,
		mature INTEGER NOT NULL,
		thumbnailIdx INTEGER NOT NULL,
		-- settings
		useDefaultsBitfield INTEGER NOT NULL,
		fetchOnStartup INTEGER NOT NULL, --bool
		intervalFetching INTEGER NOT NULL, --bool
		fetchFrequency INTEGER NOT NULL,
		notifMode INTEGER NOT NULL,
		accumCount INTEGER NOT NULL,
		delayDuration INTEGER NOT NULL,
		downloadsPath TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS titles(
		id INTEGER PRIMARY KEY,
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		title TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS titles_cid_idx On titles(comicId);
	CREATE TABLE IF NOT EXISTS thumbnails(
		id INTEGER PRIMARY KEY,
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		thumbnail TEXT NOT NULL
	);
	CREATE INDEX IF NOT EXISTS thumbnails_cid_idx On thumbnails(comicId);
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
		CONSTRAINT pk_CI_TA PRIMARY KEY (comicId, tagId)
	);
	` + SQLSourceLinkSchema() + SQLChapterSchema()
}

type syncRWMutex struct {
	internal sync.RWMutex
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

type Comic struct {
	info   ComicInfo
	config ComicConfig

	chapters        map[ChapterIdentity]*Chapter
	chaptersOrder   ChapterIdentitiesSlice
	lastReadChapter struct {
		valid    bool
		identity ChapterIdentity
	}
	cachedReadCount int

	links     []SourceLink
	preferred comicPrefs

	loadingFromDB bool
	sqlId         int64

	lock sync.RWMutex
}
type comicPrefs struct {
	sources    map[SourceId]linkIdx
	scanlators map[JointScanlatorIds]jointIdx
	color      bool
	cacheKey   cacheKey
}
type (
	linkIdx  int
	jointIdx int
	cacheKey int
)

func NewComic(cfg ComicConfig) Comic {
	return Comic{
		config:   cfg,
		chapters: make(map[ChapterIdentity]*Chapter, 128),
		preferred: comicPrefs{
			sources:    make(map[SourceId]linkIdx, 4),
			scanlators: make(map[JointScanlatorIds]jointIdx, 8),
		},
	}
}

func newComicFromDB(cfg ComicConfig) Comic {
	c := NewComic(cfg)
	c.loadingFromDB = true
	return c
}

func (this *Comic) loadingDone() {
	this.loadingFromDB = false
}

func (this *Comic) Info() ComicInfo {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.info
}

func (this *Comic) SetInfo(info ComicInfo) { //TODO: incremental insert
	this.lock.Lock()
	defer this.lock.Unlock()
	this.info = info

	if this.loadingFromDB || saveOff {
		return
	}

	db := qdb.DB()
	transaction, _ := db.Begin()
	defer transaction.Rollback()

	var newId int64
	result, err := comicInsertion.Exec(
		this.SQLId(),
		this.info.MainTitleIdx, this.info.Type, this.info.Status, this.info.ScanlationStatus, this.info.Description,
		this.info.Rating, this.info.Mature, this.info.ThumbnailIdx,
		qutils.BoolsToBitfield(this.config.OverrideDefaults), this.config.FetchOnStartup,
		this.config.IntervalFetching, this.config.FetchFrequency, this.config.NotificationMode,
		this.config.AccumulativeModeCount, this.config.DelayedModeDuration,
	)
	if err != nil {
		qlog.Logf(qlog.Error, "Error while inserting info for comic %s: %v", info.Titles[info.MainTitleIdx], err)
		return
	}
	newId, _ = result.LastInsertId()
	atomic.StoreInt64(&this.sqlId, newId)

	for _, title := range this.info.Titles {
		_, err = titleInsertion.Exec(newId, title)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting title for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}

	for _, author := range this.info.Authors {
		authorRelation.Exec(newId, author)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting author for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}
	for _, artist := range this.info.Artists {
		artistRelation.Exec(newId, artist)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting artist for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}
	for genre := range this.info.Genres {
		genreRelation.Exec(newId, genre)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting genre for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}
	for tag := range this.info.Categories {
		tagRelation.Exec(newId, tag)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting tag for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}
	for thumbnail := range this.info.Thumbnails {
		thumbnailInsertion.Exec(0, newId, thumbnail)
		if err != nil {
			qlog.Logf(qlog.Error, "Error while inserting thumbnail for comic %s: %v", info.Titles[info.MainTitleIdx], err)
			return
		}
	}

	transaction.Commit()
}

func (this *Comic) Config() ComicConfig {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.config
}

func (this *Comic) SetConfig(cfg ComicConfig) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.config = cfg

	if this.loadingFromDB {
		return
	}

	var newId int64
	result, err := comicInsertion.Exec(
		this.SQLId(),
		this.info.MainTitleIdx, this.info.Type, this.info.Status, this.info.ScanlationStatus, this.info.Description,
		this.info.Rating, this.info.Mature, this.info.ThumbnailIdx,
		qutils.BoolsToBitfield(this.config.OverrideDefaults), this.config.FetchOnStartup,
		this.config.IntervalFetching, this.config.FetchFrequency, this.config.NotificationMode,
		this.config.AccumulativeModeCount, this.config.DelayedModeDuration,
	)
	if err != nil {
		qlog.Logf(qlog.Error, "Error while inserting info for comic %s: %v", this.info.Titles[this.info.MainTitleIdx], err)
		return
	}
	newId, _ = result.LastInsertId()
	atomic.StoreInt64(&this.sqlId, newId)
}

func (this *Comic) AddSourceLink(link SourceLink) (alreadyAdded bool) {
	return this.AddSourceLinkAt(len(this.links), link)
}

func (this *Comic) AddSourceLinkAt(index int, link SourceLink) (alreadyAdded bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	lidx, alreadyAdded := this.preferred.sources[link.SourceId]
	index = int(lidx)
	if alreadyAdded {
		link.sqlId = this.links[index].sqlId //copy sqlId, so we can properly insert it into the DB
		this.links[index] = link
	} else {
		if index < len(this.links) { //insert
			this.links = append(this.links, SourceLink{})
			copy(this.links[index+1:], this.links[index:])
			this.links[index] = link
		} else { //append
			this.links = append(this.links, link)
		}
		this.preferred.sources[link.SourceId] = linkIdx(index)
		this.preferred.cacheKey++
	}

	if !this.loadingFromDB {
		err := link.sqlInsert(this.SQLId())
		if err != nil {
			qlog.Logf(qlog.Error, "Error inserting source link for comic %s: %v", this.info.Titles[this.info.MainTitleIdx], err)
		}
	}

	return alreadyAdded
}

func (this *Comic) RemoveSource(link SourceLink) (success bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	index, exists := this.preferred.sources[link.SourceId]
	if exists {
		this.links = append(this.links[:index], this.links[index+1:]...)
		delete(this.preferred.sources, link.SourceId)
		for sourceId, idx := range this.preferred.sources {
			if idx > index {
				this.preferred.sources[sourceId] = idx - 1
			}
		}
		this.preferred.cacheKey++

		linkDeletion.Exec(link.sqlId)
	}
	return exists
}

func (this *Comic) SourceLinks() []SourceLink {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.links
}

func (this *Comic) PreferredSources() map[SourceId]linkIdx {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.preferred.sources
}

func (this *Comic) AddChapter(identity ChapterIdentity, chapter Chapter) (merged bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	existingChapter, merged := this.chapters[identity]
	if chapter.MarkedRead && this.lastReadChapter.identity.LessEq(identity) {
		this.lastReadChapter.identity = identity
		this.lastReadChapter.valid = true
	}
	if merged {
		existingRead := existingChapter.MarkedRead
		existingChapter.MergeWith(&chapter)
		//this.chapters[identity] = existingChapter //reinsert
		if !existingRead && existingChapter.MarkedRead {
			this.cachedReadCount += 1
		} else if existingRead && !existingChapter.MarkedRead {
			this.cachedReadCount -= 1
		}
	} else {
		chapter.setParent(this)
		//this.preferred.scanlators = qutils.AppendUnique(this.preferred.scanlators, chapter.Scanlators()).([]JointScanlatorIds)
		//this.preferred.cacheKey++
		this.chapters[identity] = &chapter //todo: copy to memory arena first
		this.chaptersOrder.Insert(this.chaptersOrder.fittingIndexOf(identity), identity)

		chapter.sqlInsert(this, identity)

		if chapter.MarkedRead {
			this.cachedReadCount += 1
		}
	}
	return
}

func (this *Comic) AddMultipleChapters(identities []ChapterIdentity, chapters []Chapter) { //todo: merge slice types into []IdentifiedChapter to improve cache locality
	this.lock.Lock()
	defer this.lock.Unlock()

	db := qdb.DB()
	transaction, _ := db.Begin()
	defer transaction.Rollback()

	if len(identities) != len(chapters) {
		qlog.Log(qlog.Warning, "Comic.AddMultipleChapters: provided slices lengths do not match!")
	}
	minLen := math.Min(len(identities), len(chapters))
	separatedNewIdentities := make([][]ChapterIdentity, 0, minLen/2) //Slice of slices of new identities
	newSliceStartsAt := 0                                            //Starting index of new slice of new identities
	sliceOfNewEnds := false                                          //Status of creation of the slice
	for i := 0; i < minLen; i++ {
		identity := identities[i]
		chapter := chapters[i]
		existingChapter, exists := this.chapters[identity]
		if chapter.MarkedRead && this.lastReadChapter.identity.LessEq(identity) {
			this.lastReadChapter.identity = identity
			this.lastReadChapter.valid = true
		}
		if exists {
			existingRead := existingChapter.MarkedRead
			existingChapter.MergeWith(&chapter)
			if sliceOfNewEnds { //Sequence ended, add newly created slice to the list
				separatedNewIdentities = append(separatedNewIdentities, identities[newSliceStartsAt:i])
				sliceOfNewEnds = false
			}
			//this.chapters[identity] = existingChapter //reinsert
			if !existingRead && existingChapter.MarkedRead {
				this.cachedReadCount += 1
			} else if existingRead && !existingChapter.MarkedRead {
				this.cachedReadCount -= 1
			}
		} else {
			chapter.setParent(this)
			//this.preferred.scanlators = qutils.AppendUnique(this.preferred.scanlators, chapter.Scanlators()).([]JointScanlatorIds)
			//this.preferred.cacheKey++
			this.chapters[identity] = &chapter //todo: copy to memory arena first, otherwise the pointers will prevent the slice from being garbage collected
			if chapter.MarkedRead {
				this.cachedReadCount += 1
			}

			chapter.sqlInsert(this, identity)

			if !sliceOfNewEnds { //Sequence started, set starting index
				newSliceStartsAt = i
				sliceOfNewEnds = true
			}
		}
	}
	if sliceOfNewEnds { //Sequence ended
		separatedNewIdentities = append(separatedNewIdentities, identities[newSliceStartsAt:])
	}

	for _, newIdentities := range separatedNewIdentities {
		insertionIndex := int(this.chaptersOrder.fittingIndexOf(newIdentities[0]))
		this.chaptersOrder.InsertMultiple(insertionIndex, newIdentities)
	}
}

func (this *Comic) GetChapter(index int) (Chapter, ChapterIdentity) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	identity := this.chaptersOrder[index]
	return *this.chapters[identity], identity
}

func (this *Comic) Chapter(index int) (*Chapter, ChapterIdentity) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	identity := this.chaptersOrder[index]
	return this.chapters[identity], identity
}

func (this *Comic) PreferredScanlators() map[JointScanlatorIds]jointIdx {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.preferred.scanlators
}

func (this *Comic) SetPreferredScanlators(priority []JointScanlatorIds) { //TODO: use this
	this.lock.Lock()
	defer this.lock.Unlock()
	this.preferred.scanlators = make(map[JointScanlatorIds]jointIdx, len(priority))
	for i, joint := range priority {
		this.preferred.scanlators[joint] = jointIdx(i)
	}
	this.preferred.cacheKey++
}

func (this *Comic) Scanlators() []JointScanlatorIds { //will be probably slow as fuck, but that's okay
	this.lock.Lock()
	defer this.lock.Unlock()

	unique := make(map[JointScanlatorIds]struct{}, len(this.chaptersOrder)/3) //honest guess
	for _, chapter := range this.chapters {
		for _, joint := range chapter.Scanlators() {
			unique[joint] = struct{}{}
		}
	}
	ret := make([]JointScanlatorIds, len(unique))
	for joint := range unique {
		ret = append(ret, joint)
	}
	return ret //unsorted
}

func (this *Comic) ChaptersCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return len(this.chaptersOrder)
}

func (this *Comic) ChaptersReadCount() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	return this.cachedReadCount
}

func (this *Comic) LastReadChapter() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	idx, err := qutils.IndexOf(this.chaptersOrder, this.lastReadChapter.identity)
	if err != nil {
		return 0
	}
	return idx
}

func (this *Comic) QueuedChapter() int {
	this.lock.RLock()
	defer this.lock.RUnlock()
	idx, err := qutils.IndexOf(this.chaptersOrder, this.lastReadChapter.identity)
	if err != nil {
		return 0
	}
	idx++
	clen := len(this.chaptersOrder)
	if idx < clen {
		return idx
	}
	return clen - 1
}

func (this *Comic) SetColorPreference(prefersColor bool) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.preferred.color = prefersColor
	this.preferred.cacheKey++
}

func (this *Comic) PrefersColor() bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.preferred.color
}

func (this *Comic) preferencesStale(key cacheKey) bool {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.preferred.cacheKey > key
}

func (this *Comic) preferencesCacheKey() cacheKey {
	this.lock.Lock()
	defer this.lock.Unlock()
	return this.preferred.cacheKey
}

func (this *Comic) SQLId() int64 {
	return atomic.LoadInt64(&this.sqlId)
}

func (this *Comic) SQLInsert() (err error) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, identity := range this.chaptersOrder {
		//chapter := this.chapters[identity] //can't take a pointer
		//err = chapter.SQLInsert(identity, stmts)
		err = this.chapters[identity].sqlInsert(this, identity)
		if err != nil {
			return qerr.NewLocated(err)
		}
		//this.chapters[identity] = chapter //so reinsert
	}

	return nil
}

func SQLComicQuery(rows *sql.Rows) (Comic, error) {
	comic := newComicFromDB(ComicConfig{})
	info := &comic.info //todo: this is a data race
	cfg := &comic.config
	var comicId int64
	var overrideDefaultsBitfield uint64
	err := rows.Scan(
		&comicId,
		&info.MainTitleIdx, &info.Type, &info.Status, &info.ScanlationStatus, &info.Description, &info.Rating, &info.Mature, &info.ThumbnailIdx,
		&overrideDefaultsBitfield, &cfg.FetchOnStartup, &cfg.IntervalFetching, &cfg.FetchFrequency, &cfg.NotificationMode,
		&cfg.AccumulativeModeCount, &cfg.DelayedModeDuration,
	)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	atomic.StoreInt64(&comic.sqlId, comicId)
	cfg.OverrideDefaults = qutils.BitfieldToBools(overrideDefaultsBitfield, bitlength_comiccfg)

	titleRows, err := titlesQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for titleRows.Next() {
		var title string
		err = titleRows.Scan(&title)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Titles = append(info.Titles, title)
	}

	authorRows, err := authorsQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for authorRows.Next() {
		var author AuthorId
		err = authorRows.Scan(&author)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Authors = append(info.Authors, author)
	}

	artistRows, err := artistsQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for artistRows.Next() {
		var artist ArtistId
		err = artistRows.Scan(&artist)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Artists = append(info.Artists, artist)
	}

	genreRows, err := genresQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for genreRows.Next() {
		var genre ComicGenreId
		err = genreRows.Scan(&genre)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Genres = append(info.Genres, genre)
	}

	tagRows, err := tagsQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for tagRows.Next() {
		var tag ComicTagId
		err = tagRows.Scan(&tag)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Categories = append(info.Categories, tag)
	}

	thumbnailRows, err := thumbnailsQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for thumbnailRows.Next() {
		var thumbnail string
		err = thumbnailRows.Scan(&thumbnail)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		info.Thumbnails = append(info.Thumbnails, thumbnail)
	}

	sourceRows, err := sourcesQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for sourceRows.Next() {
		source, err := SQLSourceLinkQuery(sourceRows)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		comic.AddSourceLink(source)
	}

	var identities []ChapterIdentity
	var chapters []Chapter
	chapterRows, err := chaptersQuery.Query(comicId)
	if err != nil {
		return Comic{}, qerr.NewLocated(err)
	}
	for chapterRows.Next() {
		chapter, identity, err := SQLChapterQuery(chapterRows)
		if err != nil {
			return Comic{}, qerr.NewLocated(err)
		}
		identities = append(identities, identity)
		chapters = append(chapters, chapter)
	}
	comic.AddMultipleChapters(identities, chapters)

	comic.loadingDone()
	return comic, nil
}

type ComicInfo struct {
	MainTitleIdx     int
	Titles           []string
	Authors          []AuthorId
	Artists          []ArtistId
	Genres           []ComicGenreId
	Categories       []ComicTagId
	Type             comicType
	Status           comicStatus
	ScanlationStatus ScanlationStatus
	Description      string
	Rating           uint16
	Mature           bool
	ThumbnailIdx     int
	Thumbnails       []string
}

func (this ComicInfo) MergeWith(another *ComicInfo) (merged ComicInfo) {
	this.Titles = qutils.AppendUnique(this.Titles, another.Titles).([]string)
	this.Authors = qutils.AppendUnique(this.Authors, another.Authors).([]AuthorId)
	this.Artists = qutils.AppendUnique(this.Artists, another.Artists).([]ArtistId)
	this.Genres = qutils.AppendUnique(this.Genres, another.Genres).([]ComicGenreId)
	this.Categories = qutils.AppendUnique(this.Categories, another.Categories).([]ComicTagId)

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

	this.Rating = uint16(math.Max(int(this.Rating), int(another.Rating)))
	this.Mature = another.Mature || this.Mature
	this.Thumbnails = qutils.AppendUnique(this.Thumbnails, another.Thumbnails).([]string)

	return this
}
