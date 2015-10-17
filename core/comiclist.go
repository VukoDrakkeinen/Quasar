package core

import (
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"math"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"
)

type ViewNotificationType int

const (
	Insert ViewNotificationType = iota
	Remove
	Reset
	Update
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
	scheduleSchema = `
	CREATE TABLE IF NOT EXISTS schedule(
		comicId INTEGER PRIMARY KEY,
		nextFetchTime TIMESTAMP NOT NULL,
		lastUpdated TIMESTAMP NOT NULL
	);
	`
	idsInsertionPreCmd   = `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	idsQueryPreCmd       = `SELECT $colName FROM $tableName;`
	scheduleInsertionCmd = `INSERT OR REPLACE INTO schedule(comicId, nextFetchTime, lastUpdated)
							VALUES((SELECT id FROM comics WHERE id = ?), ?, ?);`
	scheduleQueryCmd = `SELECT nextFetchTime, lastUpdated FROM schedule WHERE comicId = ?;`
)

func NewComicList(fetcher *fetcher, notifyViewFunc func(typ ViewNotificationType, row, count int, work func())) *ComicList {
	if notifyViewFunc == nil {
		notifyViewFunc = func(a ViewNotificationType, b, c int, work func()) {
			work()
		}
	}

	return &ComicList{
		comics:     make([]*Comic, 0, 10),
		metadata:   make([]comicMetadata, 0, 10),
		fetcher:    fetcher,
		notifyView: notifyViewFunc,
	}
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
	comics     []*Comic
	metadata   []comicMetadata
	fetcher    *fetcher
	notifyView func(typ ViewNotificationType, row, count int, work func())
	dataLock   sync.RWMutex
	metaLock   sync.RWMutex
}

func (list ComicList) Fetcher() *fetcher {
	return list.fetcher
}

func (this *ComicList) AddComics(comics ...*Comic) {
	this.notifyView(Insert, len(this.comics), len(comics), func() {
		this.dataLock.Lock()
		this.metaLock.Lock()
		defer this.dataLock.Unlock()
		defer this.metaLock.Unlock()

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
	})
}

func (this *ComicList) RemoveComics(index, count int) {
	this.notifyView(Remove, int(index), int(count), func() {
		this.dataLock.Lock()
		this.metaLock.Lock()
		defer this.dataLock.Unlock()
		defer this.metaLock.Unlock()

		this.comics = this.comics[:index+copy(this.comics[index:], this.comics[index+count:])]
		for i := index; i < (index + count); i++ {
			this.cancelScheduleForComic(i)
		}

		this.metadata = this.metadata[:index+copy(this.metadata[index:], this.metadata[index+count:])]
		for i := index + count; i < len(this.metadata); i++ {
			this.scheduleComicFetchFor(i, true)
		}
	})
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
	return this.comics[idx]
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
		if fetchOnStartup && !reschedule {
			this.metadata[comicIdx].nextFetchAt = time.Time{} //FIXME: this is hack, but I'm too fed up with this piece of code to fix it
			this.comicFetch(comicIdx, false, false)
		} else if intervalFetching && this.metadata[comicIdx].nextFetchAt.Before(time.Now().UTC()) {
			this.comicFetch(comicIdx, true, false)
		}

		if intervalFetching {
			for {
				select {
				case <-time.After(this.metadata[comicIdx].nextFetchAt.Sub(time.Now().UTC())):
					this.comicFetch(comicIdx, false, false)
				case <-this.metadata[comicIdx].schedInterrupt:
					return
				}
			}
		}
	}()

}

func (this ComicList) UpdateComic(comicId int) {
	this.comicFetch(comicId, false, true)
}

func (this ComicList) comicFetch(comicIdx int, missedFetches, manual bool) {
	now := time.Now().UTC()
	canUpdate := true
	this.notifyView(Update, comicIdx, 1, func() {
		canUpdate = atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(&this.metadata[comicIdx].status)),
			uint32(ComicNotUpdating), uint32(ComicUpdating),
		)
	})
	if !canUpdate {
		return
	}

	if manual {
		this.dataLock.Lock()
		this.cancelScheduleForComic(comicIdx)
		this.dataLock.Unlock()
	}

	this.notifyView(Update, comicIdx, 1, func() {
		comic := this.GetComic(comicIdx)
		freq := this.comicUpdateFrequency(comicIdx)

		prev := this.metadata[comicIdx].nextFetchAt
		if manual || prev.IsZero() {
			prev = now
			missedFetches = false
		}
		if !missedFetches {
			this.metadata[comicIdx].nextFetchAt = prev.Add(freq)
		} else {
			multiplier := divCeil(now.Sub(prev), freq)
			this.metadata[comicIdx].nextFetchAt = prev.Add(multiplier * freq)
		}

		this.fetcher.DownloadChapterListFor(comic)
		this.metadata[comicIdx].updatedAt = now

		atomic.StoreUint32((*uint32)(unsafe.Pointer(&this.metadata[comicIdx].status)), uint32(ComicNotUpdating))
	})

	if manual {
		this.scheduleComicFetchFor(comicIdx, true)
	}
}

func (this ComicList) comicUpdateFrequency(comicIdx int) time.Duration {
	comic := this.GetComic(comicIdx)
	cSettings := comic.Settings()
	fSettings := this.fetcher.settings

	if overrideFrequency := cSettings.OverrideDefaults[2]; overrideFrequency {
		return cSettings.FetchFrequency
	} else {
		return fSettings.FetchFrequency
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

func divCeil(divident, divisor time.Duration) (multiplier time.Duration) {
	x := float64(divident)
	y := float64(divisor)
	return time.Duration(math.Ceil(x / y))
}

func CreateDB(db *qdb.QDB) (err error) {
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
func (this ComicList) SaveToDB() {
	db := qdb.DB()
	if db == nil {
		qlog.Log(qlog.Error, "Database handle is nil! Aborting save.")
		return
	}
	err := CreateDB(db)
	if err != nil {
		qlog.Log(qlog.Error, "Creating database failed.", qerr.NewLocated(err))
		return
	}

	type tuple struct {
		dict qdb.InsertionStmtExecutor
		name string
	}
	for _, tuple := range []tuple{ //TODO?: global state is bad
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
			qlog.Log(qlog.Error, "Error while inserting into", tuple.name, "table:", err)
			return
		}
		transaction.Commit()
	}

	this.dataLock.RLock()
	this.metaLock.RLock()
	defer this.dataLock.RUnlock()
	defer this.metaLock.RUnlock()

	scheduleInsertionStmt := db.MustPrepare(scheduleInsertionCmd)
	dbStmts := SQLComicInsertStmts(db)
	defer dbStmts.Close()
	for i, comic := range this.comics {
		transaction, _ := db.Begin()

		err := comic.SQLInsert(dbStmts.ToTransactionSpecific(transaction))
		if err != nil { // no need to manually close statements, Commit() or Rollback() take care of that
			qlog.Log(qlog.Error, "Error while saving, rolling back:", qerr.NewLocated(err))
			transaction.Rollback()
			continue
		}
		_, err = transaction.Stmt(scheduleInsertionStmt).Exec(comic.sqlId, this.metadata[i].nextFetchAt, this.metadata[i].updatedAt)
		if err != nil {
			qlog.Log(qlog.Error, "Error while saving, rolling back:", qerr.NewLocated(err))
			transaction.Rollback()
			continue
		}
		transaction.Commit()
	}
}

func (list *ComicList) LoadFromDB() (err error) {
	list.cancelSchedule()

	db := qdb.DB()
	CreateDB(db)
	transaction, _ := db.Begin()
	defer transaction.Rollback()

	type tuple struct {
		dict qdb.QueryStmtExecutor
		name string
	}
	for _, tuple := range []tuple{ //FIXME: Global dicts are bad
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
			return qerr.NewLocated(err)
		}
		idsQueryStmt.Close()
	}

	idsdict.Langs.AssignIds(list.fetcher.PluginProvidedLanguages())

	list.dataLock.RLock()
	comicSqlIds := make(map[int64]struct{}) //TODO: not satisfied with this part, rewrite
	for _, comic := range list.comics {
		comicSqlIds[comic.sqlId] = struct{}{}
	}
	list.dataLock.RUnlock()

	dbStmts := SQLComicQueryStmts(db)
	defer dbStmts.Close()
	stmts := dbStmts.ToTransactionSpecific(transaction)
	scheduleQueryStmt := db.MustPrepare(scheduleQueryCmd)
	comicRows, err := stmts[comicsQuery].Query()
	if err != nil {
		return qerr.NewLocated(err)
	}

	for comicRows.Next() {
		comic, err := SQLComicQuery(comicRows, stmts)
		if err != nil {
			return qerr.NewLocated(err)
		}

		if _, exists := comicSqlIds[comic.sqlId]; exists { //TODO: don't skip, merge
			qlog.Logf(qlog.Info, "Skipped one comic while loading from DB (already exists). Id: %d", comic.sqlId)
			continue
		}

		list.dataLock.RLock()
		clen := len(list.comics)
		list.dataLock.RUnlock()
		list.notifyView(Insert, clen, 1, func() {
			list.dataLock.Lock()
			list.metaLock.Lock()
			defer list.dataLock.Unlock()
			defer list.metaLock.Unlock()

			list.comics = append(list.comics, comic)
			var nextFetchTime, lastUpdated time.Time
			err = scheduleQueryStmt.QueryRow(comic.sqlId).Scan(&nextFetchTime, &lastUpdated)
			if err != nil {
				return
			}

			list.metadata = append(list.metadata,
				comicMetadata{
					status:         ComicNotUpdating,
					updatedAt:      lastUpdated,
					nextFetchAt:    nextFetchTime.UTC(),
					schedInterrupt: make(chan struct{}),
				},
			)
		})
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	transaction.Commit()

	list.ScheduleComicFetches() //TODO: only the new ones
	return
}
