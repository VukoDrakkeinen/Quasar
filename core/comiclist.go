package core

import (
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"database/sql"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/eventq"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

var (
	saveOff = true //todo: remove this unholy abomination

	ComicsAboutToBeAdded = eventq.NewEventType()
	ComicsAdded          = eventq.NewEventType()

	ComicsAboutToBeRemoved = eventq.NewEventType()
	ComicsRemoved          = eventq.NewEventType()

	ComicsUpdateStatusChanged = eventq.NewEventType()

	ChapterListAboutToChange = eventq.NewEventType()
	ChapterListChanged       = eventq.NewEventType()
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
	scheduleSchema = `
	CREATE TABLE IF NOT EXISTS schedule(
		comicId INTEGER NOT NULL REFERENCES comics(id),
		nextFetchTime TIMESTAMP NOT NULL,
		lastUpdated TIMESTAMP NOT NULL
	);
	CREATE INDEX IF NOT EXISTS schedule_cid_idx ON schedule(comicId);
	` //TODO: db version
	idsInsertionPreCmd   = `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	idsQueryPreCmd       = `SELECT $colName FROM $tableName;`
	scheduleInsertionCmd = `INSERT OR REPLACE INTO schedule(comicId, nextFetchTime, lastUpdated)
							VALUES((SELECT id FROM comics WHERE id = ?), ?, ?);`
	scheduleQueryCmd = `SELECT nextFetchTime, lastUpdated FROM schedule WHERE comicId = ?;`

	scheduleQuery *sql.Stmt

	langIdInsertion      *sql.Stmt
	scanlatorIdInsertion *sql.Stmt
	authorIdInsertion    *sql.Stmt
	artistIdInsertion    *sql.Stmt
	genreIdInsertion     *sql.Stmt
	tagIdInsertion       *sql.Stmt
)

func init() {
	qdb.PrepareStmt(&scheduleQuery, scheduleQueryCmd)

	type tuple struct {
		assignToVar **sql.Stmt
		name        string
	}
	for _, tuple := range []tuple{
		{&langIdInsertion, "lang"},
		{&scanlatorIdInsertion, "scanlator"},
		{&authorIdInsertion, "author"},
		{&artistIdInsertion, "artist"},
		{&genreIdInsertion, "genre"},
		{&tagIdInsertion, "tag"},
	} {
		rep := strings.NewReplacer("$tableName", tuple.name+"s", "$colName", tuple.name)
		qdb.PrepareStmt(tuple.assignToVar, rep.Replace(idsInsertionPreCmd))
	}
}

func NewComicList(fetcher fetcher) *ComicList {
	list := &ComicList{
		comics:     make([]Comic, 0, 10),
		metadata:   make([]comicMetadata, 0, 10),
		fetcher:    fetcher,
		langs:      idsdict.NewLangDict(),
		scanlators: idsdict.NewScanlatorsDict(),
		authors:    idsdict.NewAuthorDict(),
		artists:    idsdict.NewArtistsDict(),
		genres:     idsdict.NewComicGenresDict(),
		tags:       idsdict.NewComicTagsDict(),

		Messenger: eventq.NewMessenger(),
	}

	if !saveOff {
		list.langs.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			langIdInsertion.Exec(int(id) + 1)
		})
		list.scanlators.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			scanlatorIdInsertion.Exec(int(id) + 1)
		})
		list.authors.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			authorIdInsertion.Exec(int(id) + 1)
		})
		list.artists.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			artistIdInsertion.Exec(int(id) + 1)
		})
		list.genres.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			genreIdInsertion.Exec(int(id) + 1)
		})
		list.tags.On(idsdict.IdAssigned).Do(func(args ...interface{}) {
			id := args[0].(idsdict.Id)
			tagIdInsertion.Exec(int(id) + 1)
		})
	}
	return list
}

const (
	ComicNotUpdating ComicUpdatingEnum = iota
	ComicUpdating
)

type comicMetadata struct {
	status         ComicUpdatingEnum
	updatedAt      time.Time
	nextFetchAt    time.Time
	schedInterrupt chan struct{}
}

type ComicUpdatingEnum uint32
type ComicList struct {
	comics   []Comic
	metadata []comicMetadata
	fetcher  fetcher
	dataLock sync.RWMutex
	metaLock sync.RWMutex

	langs      idsdict.LangsDict
	scanlators idsdict.ScanlatorsDict
	authors    idsdict.AuthorsDict
	artists    idsdict.ArtistsDict
	genres     idsdict.ComicGenresDict
	tags       idsdict.ComicTagsDict

	eventq.Messenger
}

func (list *ComicList) Fetcher() *fetcher {
	return &list.fetcher
}

func (list *ComicList) Langs() *idsdict.LangsDict {
	return &list.langs
}

func (list *ComicList) Scanlators() *idsdict.ScanlatorsDict {
	return &list.scanlators
}

func (list *ComicList) Authors() *idsdict.AuthorsDict {
	return &list.authors
}

func (list *ComicList) Artists() *idsdict.ArtistsDict {
	return &list.artists
}

func (list *ComicList) Genres() *idsdict.ComicGenresDict {
	return &list.genres
}

func (list *ComicList) Tags() *idsdict.ComicTagsDict {
	return &list.tags
}

func (this *ComicList) AddComics(comics ...Comic) {
	this.dataLock.Lock()
	this.metaLock.Lock()
	this.Event(ComicsAboutToBeAdded, len(this.comics), len(comics))
	defer this.dataLock.Unlock()
	defer this.metaLock.Unlock()
	defer this.Event(ComicsAdded)

	this.comics = append(this.comics, comics...)
	mlen := len(this.metadata)
	newLen := mlen + len(comics)

	if cap(this.metadata) < newLen {
		metadata := make([]comicMetadata, newLen, qutils.GrownCap(newLen))
		copy(metadata, this.metadata)
		this.metadata = metadata
	} else {
		this.metadata = this.metadata[:newLen]
	}

	for i := range this.metadata[mlen:] {
		this.metadata[mlen+i].schedInterrupt = make(chan struct{})
	}
}

func (this *ComicList) RemoveComics(index, count int) {
	this.dataLock.Lock()
	this.metaLock.Lock()
	this.Event(ComicsAboutToBeRemoved, index, count)
	defer this.dataLock.Unlock()
	defer this.metaLock.Unlock()
	defer this.Event(ComicsRemoved)

	this.comics = this.comics[:index+copy(this.comics[index:], this.comics[index+count:])]
	for i := index; i < (index + count); i++ {
		this.cancelScheduleForComic(i)
	}

	this.metadata = this.metadata[:index+copy(this.metadata[index:], this.metadata[index+count:])]
	for i := index + count; i < len(this.metadata); i++ {
		this.scheduleComicFetchFor(i, true)
	}
}

func (this ComicList) ComicLastUpdated(idx int) time.Time {
	this.metaLock.RLock()
	defer this.metaLock.RUnlock()
	return this.metadata[idx].updatedAt
}

func (this ComicList) ComicIsUpdating(idx int) bool {
	this.metaLock.Lock()
	defer this.metaLock.Unlock()
	return this.metadata[idx].status == ComicUpdating
}

func (this ComicList) GetComic(idx int) *Comic {
	this.dataLock.Lock()
	defer this.dataLock.Unlock()
	return &this.comics[idx]
}

func (this *ComicList) Len() int {
	this.dataLock.RLock()
	defer this.dataLock.RUnlock()
	return len(this.comics)
}

func (this ComicList) ScheduleComicFetches() {
	for i := range this.comics {
		this.scheduleComicFetchFor(i, false)
	}
}

func (this ComicList) scheduleComicFetchFor(comicIdx int, reschedule bool) {
	this.dataLock.Lock()
	this.cancelScheduleForComic(comicIdx)
	this.dataLock.Unlock()

	fetchOnStartup := this.fetcher.settings.FetchOnStartup
	intervalFetching := this.fetcher.settings.IntervalFetching

	go func() {
		return //todo: temp, remove
		if fetchOnStartup && !reschedule {
			this.comicFetch(Startup, comicIdx)
		}

		if intervalFetching {
			for {
				select {
				case <-time.After(this.metadata[comicIdx].nextFetchAt.Sub(time.Now().UTC())):
					this.comicFetch(Scheduled, comicIdx)
				case <-this.metadata[comicIdx].schedInterrupt:
					return
				}
			}
		}
	}()

}

func (this ComicList) UpdateComic(comicId int) {
	this.comicFetch(Manual, comicId)
}

type fetchType int

const (
	Scheduled fetchType = iota
	Startup
	Manual
)

var unixEpoch = time.Unix(0, 0)

func (this ComicList) comicFetch(fetchType fetchType, comicIdx int) {
	if fetchType == Startup && comicIdx != 88 { //FIXME: test
		return
	}
	now := time.Now().UTC()
	canUpdate := atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(&this.metadata[comicIdx].status)),
		uint32(ComicNotUpdating), uint32(ComicUpdating),
	)
	this.Event(ComicsUpdateStatusChanged, comicIdx, 1)
	if !canUpdate {
		return
	}

	if fetchType == Manual {
		this.dataLock.Lock()
		this.cancelScheduleForComic(comicIdx)
		this.dataLock.Unlock()
	}

	comic := this.GetComic(comicIdx)
	freq := this.comicUpdateFrequency(comicIdx)

	prev := this.metadata[comicIdx].nextFetchAt
	next := prev
	switch fetchType {
	case Startup:
		if prev.Before(unixEpoch) {
			next = now.Add(freq)
			break
		}
		fallthrough
	case Scheduled:
		next = prev.Add(freq * divCeil(now.Sub(prev), freq))
	case Manual:
		next = now.Add(freq)
	}
	this.metadata[comicIdx].nextFetchAt = next

	this.Event(ChapterListAboutToChange) //todo: do only when comicIdx == chapterView.comicId
	this.fetcher.FetchChapterListFor(comic)
	this.Event(ChapterListChanged)
	this.metadata[comicIdx].updatedAt = now

	atomic.StoreUint32((*uint32)(unsafe.Pointer(&this.metadata[comicIdx].status)), uint32(ComicNotUpdating))
	this.Event(ComicsUpdateStatusChanged, comicIdx, 1)

	if fetchType == Manual {
		this.scheduleComicFetchFor(comicIdx, true)
	}
}

func (this ComicList) comicUpdateFrequency(comicIdx int) time.Duration {
	comic := this.GetComic(comicIdx)
	cSettings := comic.Config()
	fSettings := this.fetcher.settings

	if overrideFrequency := cSettings.OverrideDefaults[2]; overrideFrequency {
		return time.Duration(cSettings.FetchFrequency)
	} else {
		return time.Duration(fSettings.FetchFrequency)
	}
}

func (this *ComicList) cancelSchedule() {
	this.metaLock.Lock()
	defer this.metaLock.Unlock()
	for i := range this.comics {
		this.cancelScheduleForComic(i)
	}
}

func (this *ComicList) cancelScheduleForComic(comicId int) { //How?!
	close(this.metadata[comicId].schedInterrupt)                //DATA RACE: read
	this.metadata[comicId].schedInterrupt = make(chan struct{}) //DATA RACE: write
}

func divCeil(dividend, divisor time.Duration) (multiplier time.Duration) {
	return (dividend + divisor - 1) / divisor
}

func CreateDB(db *qdb.QDB) (err error) {
	//TODO: error out on nil db
	transaction, _ := db.Begin()
	defer transaction.Rollback()
	_, err = transaction.Exec(idsSchema)
	if err != nil {
		return qerr.NewLocated(err)
	}
	_, err = transaction.Exec(scheduleSchema)
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

//TODO: more error checking
//TODO: write some unit tests
//TODO: return errors, so we can show the user a pop-up
//func (this ComicList) SaveToDB() { //TODO: lock during the whole function execution
//	db := qdb.DB()
//	if db == nil {
//		qlog.Log(qlog.Error, "Database handle is nil! Aborting save.")
//		return
//	}
//	err := CreateDB(db)
//	if err != nil {
//		qlog.Log(qlog.Error, "Creating database failed.", qerr.NewLocated(err))
//		return
//	}
//
//	type tuple struct {
//		dict qdb.InsertionStmtExecutor
//		name string
//	}
//	for _, tuple := range []tuple{
//		{&this.langs, "lang"},
//		{&this.scanlators, "scanlator"},
//		{&this.authors, "author"},
//		{&this.artists, "artist"},
//		{&this.genres, "genre"},
//		{&this.tags, "tag"},
//	} {
//		func() {
//			transaction, _ := db.Begin()
//			defer transaction.Rollback()
//			rep := strings.NewReplacer("$tableName", tuple.name + "s", "$colName", tuple.name)
//			idsInsertionStmt, _ := transaction.Prepare(rep.Replace(idsInsertionPreCmd))
//			err = tuple.dict.ExecuteInsertionStmt(idsInsertionStmt)
//			if err != nil {
//				qlog.Log(qlog.Error, "Error while inserting into", tuple.name, "table:", err)
//				return
//			}
//			transaction.Commit()
//		}()
//	}
//
//	this.dataLock.RLock()
//	this.metaLock.RLock()
//	defer this.dataLock.RUnlock()
//	defer this.metaLock.RUnlock()
//
//	scheduleInsertionStmt := db.MustPrepare(scheduleInsertionCmd)
//	for i, comic := range this.comics {
//		func() {
//			transaction, _ := db.Begin()
//			defer transaction.Rollback()
//
//			err := comic.SQLInsert()
//			if err != nil {
//				// no need to manually close statements, Commit() or Rollback() take care of that
//				qlog.Log(qlog.Error, "Error while saving, rolling back:", qerr.NewLocated(err))
//				return
//			}
//			_, err = transaction.Stmt(scheduleInsertionStmt).Exec(comic.sqlId, this.metadata[i].nextFetchAt, this.metadata[i].updatedAt)
//			if err != nil {
//				qlog.Log(qlog.Error, "Error while saving, rolling back:", qerr.NewLocated(err))
//				return
//			}
//			transaction.Commit()
//		}()
//	}
//}

func (list *ComicList) LoadFromDB() (err error) { //TODO: lock during the whole function execution
	list.cancelSchedule()

	db := qdb.DB()
	err = CreateDB(db)
	if err != nil {
		return qerr.NewLocated(err)
	}

	transaction, _ := db.Begin()
	defer transaction.Rollback()

	type tuple struct {
		dict qdb.QueryStmtExecutor
		name string
	}
	for _, tuple := range []tuple{
		{&list.langs, "lang"},
		{&list.scanlators, "scanlator"},
		{&list.authors, "author"},
		{&list.artists, "artist"},
		{&list.genres, "genre"},
		{&list.tags, "tag"},
	} {
		rep := strings.NewReplacer("$tableName", tuple.name+"s", "$colName", tuple.name)
		idsQueryStmt, _ := transaction.Prepare(rep.Replace(idsQueryPreCmd))
		err := tuple.dict.ExecuteQueryStmt(idsQueryStmt)
		if err != nil {
			return qerr.NewLocated(err)
		}
		idsQueryStmt.Close()
	}

	list.langs.AssignIds(list.fetcher.PluginProvidedLanguages())

	list.dataLock.RLock()
	comicSqlIds := make(map[int64]struct{}) //TODO: not satisfied with this part, rewrite
	for _, comic := range list.comics {
		comicSqlIds[comic.sqlId] = struct{}{}
	}
	list.dataLock.RUnlock()

	comicRows, err := comicsQuery.Query()
	if err != nil {
		return qerr.NewLocated(err)
	}

	for comicRows.Next() {
		comic, err := SQLComicQuery(comicRows)
		if err != nil {
			return qerr.NewLocated(err)
		}

		if _, exists := comicSqlIds[comic.sqlId]; exists { //TODO: don't skip, merge
			qlog.Logf(qlog.Info, "Skipped one comic while loading from DB (already exists). Id: %d", comic.sqlId)
			continue
		}

		list.Event(ComicsAboutToBeAdded, len(list.comics), 1)
		list.dataLock.Lock()
		list.metaLock.Lock()

		list.comics = append(list.comics, comic) //FIXME: raw access
		var nextFetchTime, lastUpdated time.Time
		err = scheduleQuery.QueryRow(comic.sqlId).Scan(&nextFetchTime, &lastUpdated)
		if err != nil {
			return qerr.NewLocated(err)
		}

		list.metadata = append(list.metadata,
			comicMetadata{
				status:         ComicNotUpdating,
				updatedAt:      lastUpdated,
				nextFetchAt:    nextFetchTime.UTC(),
				schedInterrupt: make(chan struct{}),
			},
		)
		list.dataLock.Unlock()
		list.metaLock.Unlock()
		list.Event(ComicsAdded)
	}

	transaction.Commit()

	list.ScheduleComicFetches() //TODO: only the new ones
	return
}
