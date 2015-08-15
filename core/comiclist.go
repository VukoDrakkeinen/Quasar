package core

import (
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"math"
	"strings"
	"sync/atomic"
	"time"
	"unsafe"
)

type ViewNotificationType int

const (
	Insert ViewNotificationType = iota
	Remove
	Reset
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
	idsQueryPreCmd       = `SELECT $colName FROM $tableName;` //TODO?: use placeholders?
	scheduleInsertionCmd = `INSERT OR REPLACE INTO schedule(comicId, nextFetchTime, lastUpdated)
							VALUES((SELECT id FROM comics WHERE id = ?), ?, ?);`
	scheduleQueryCmd = `SELECT nextFetchTime, lastUpdated FROM schedule WHERE comicId = ?;`
)

func NewComicList(fetcher *fetcher, notifyViewFunc func(typ ViewNotificationType, row, count int, work func())) ComicList {
	return ComicList{
		comics:         make([]*Comic, 0, 10),
		interruptChans: make([]chan struct{}, 0, 10),
		fetcher:        fetcher,
		notifyView:     notifyViewFunc,
	}
}

const (
	ComicNotUpdating ComicUpdatingBool = iota
	ComicUpdating
)

type ComicUpdatingBool uint32
type ComicList struct {
	comics         []*Comic
	statuses       []ComicUpdatingBool
	updatedAt      []time.Time
	nextFetchTimes []time.Time
	interruptChans []chan struct{}
	fetcher        *fetcher
	notifyView     func(typ ViewNotificationType, row, count int, work func())
}

func (list ComicList) Fetcher() *fetcher {
	return list.fetcher
}

func (this *ComicList) AddComics(comics []*Comic) {
	this.notifyView(Insert, len(this.comics), len(comics), func() {
		this.comics = append(this.comics, comics...)
		interruptChans := make([]chan struct{}, len(comics))
		for i := range interruptChans {
			interruptChans[i] = make(chan struct{})
		}
		this.interruptChans = append(this.interruptChans, interruptChans...)
		this.nextFetchTimes = append(this.nextFetchTimes, make([]time.Time, len(comics))...)
		this.updatedAt = append(this.updatedAt, make([]time.Time, len(comics))...)
		this.statuses = append(this.statuses, make([]ComicUpdatingBool, len(comics))...)
	})
}

func (this *ComicList) RemoveComics(index, count int64) {
	this.notifyView(Remove, int(index), int(count), func() {
		this.comics = append(this.comics[:index], this.comics[index+count:]...)
		this.interruptChans = append(this.interruptChans[:index], this.interruptChans[index+count:]...)
		this.nextFetchTimes = append(this.nextFetchTimes[:index], this.nextFetchTimes[index+count:]...)
		this.updatedAt = append(this.updatedAt[:index], this.updatedAt[index+count:]...)
		this.statuses = append(this.statuses[:index], this.statuses[index+count:]...)
	})
}

func (this ComicList) ComicLastUpdated(idx int) time.Time {
	return this.updatedAt[idx]
}

func (this ComicList) ComicIsUpdating(idx int) bool {
	return this.statuses[idx] == ComicUpdating
}

func (this ComicList) GetComic(idx int) *Comic {
	return this.comics[idx]
}

func (this ComicList) Len() int {
	return len(this.comics)
}

func (this ComicList) ScheduleComicFetches() {
	for i := range this.comics {
		this.scheduleComicFetchFor(i, false)
	}
}

func (this ComicList) scheduleComicFetchFor(comicIdx int, reschedule bool) {
	this.cancelScheduleForComic(comicIdx)

	fetchOnStartup := this.fetcher.settings.FetchOnStartup
	intervalFetching := this.fetcher.settings.IntervalFetching

	go func() {
		//fmt.Println("  Old schedule", this.nextFetchTimes[comicIdx].Local())
		if fetchOnStartup && !reschedule {
			fmt.Println("Fetch on startup")
			this.nextFetchTimes[comicIdx] = time.Time{} //FIXME: this is hack, but I'm too fed up with this piece of code to fix it
			this.comicFetch(comicIdx, false, false)
		} else if intervalFetching && this.nextFetchTimes[comicIdx].Before(time.Now().UTC()) {
			fmt.Println("Scheduled time in the past; adjusting...")
			this.comicFetch(comicIdx, true, false)
		}

		if intervalFetching {
			for {
				select {
				case <-time.After(this.nextFetchTimes[comicIdx].Sub(time.Now().UTC())):
					fmt.Println("Scheduled fetch starting on", time.Now())
					this.comicFetch(comicIdx, false, false)
				case <-this.interruptChans[comicIdx]:
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
	this.notifyView(Reset, -1, -1, func() {
		canUpdate = atomic.CompareAndSwapUint32((*uint32)(unsafe.Pointer(&this.statuses[comicIdx])),
			uint32(ComicNotUpdating), uint32(ComicUpdating),
		)
	})
	if !canUpdate {
		fmt.Println("Cannot update while updating")
		return
	}

	if manual {
		this.cancelScheduleForComic(comicIdx)
	}

	fmt.Println("Updating")
	this.notifyView(Reset, -1, -1, func() {
		comic := this.GetComic(comicIdx)
		freq := this.comicUpdateFrequency(comicIdx)

		prev := this.nextFetchTimes[comicIdx]
		if manual || prev.IsZero() {
			prev = now
			missedFetches = false
		}
		if !missedFetches {
			this.nextFetchTimes[comicIdx] = prev.Add(freq)
		} else {
			multiplier := divCeil(now.Sub(prev), freq)
			this.nextFetchTimes[comicIdx] = prev.Add(multiplier * freq)
		}

		this.notifyView(Reset, -1, -1, func() {
			this.fetcher.DownloadChapterListFor(comic)
			this.updatedAt[comicIdx] = now
		})

		atomic.StoreUint32((*uint32)(unsafe.Pointer(&this.statuses[comicIdx])), uint32(ComicNotUpdating))
		fmt.Println(comicIdx, "Scheduled fetch for", this.nextFetchTimes[comicIdx].Local())
	})

	if manual {
		this.scheduleComicFetchFor(comicIdx, true)
	}
}

func (this ComicList) comicUpdateFrequency(comicIdx int) time.Duration {
	//return 2 * time.Minute
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
	for i := range this.comics {
		this.cancelScheduleForComic(i)
	}
}

func (this *ComicList) cancelScheduleForComic(comicId int) {
	close(this.interruptChans[comicId])
	this.interruptChans[comicId] = make(chan struct{})
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
//TODO: write a unit test
func (this ComicList) SaveToDB() {
	db := qdb.DB()
	if db == nil {
		qlog.Log(qlog.Error, "Database handle is nil! Aborting save.")
		return
		//panic()	//TODO?
	}
	err := CreateDB(db)
	if err != nil {
		qlog.Log(qlog.Error, "Creating database failed.", qerr.NewLocated(err))
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
			qlog.Log(qlog.Error, "Error while inserting into", tuple.name, "table:", err)
			return
		}
		transaction.Commit()
	}

	scheduleInsertionStmt := db.MustPrepare(scheduleInsertionCmd)
	dbStmts := SQLComicInsertStmts(db)
	defer dbStmts.Close()
	for i, comic := range this.comics {
		transaction, _ := db.Begin()
		stmts := dbStmts.ToTransactionSpecific(transaction)

		err := comic.SQLInsert(stmts)
		if err != nil { // no need to manually close statements, Commit() or Rollback() take care of that
			qlog.Log(qlog.Error, "Error while saving, rolling back:", qerr.NewLocated(err))
			transaction.Rollback()
		} else {
			transaction.Commit()
			_, err := scheduleInsertionStmt.Exec(comic.sqlId, this.nextFetchTimes[i], this.updatedAt[i]) //TODO: move to comic Tx
			if err != nil {
				qlog.Log(qlog.Error, qerr.NewLocated(err))
			}
		}
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
			return qerr.NewLocated(err)
		}
		idsQueryStmt.Close()
	}

	comicSqlIds := make(map[int64]struct{}) //TODO: not satisfied with this part, rewrite
	for _, comic := range list.comics {
		comicSqlIds[comic.sqlId] = struct{}{}
	}

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
		list.notifyView(Insert, len(list.comics), 1, func() {
			list.comics = append(list.comics, comic)
			var nextFetchTime, lastUpdated time.Time
			err = scheduleQueryStmt.QueryRow(comic.sqlId).Scan(&nextFetchTime, &lastUpdated)
			if err != nil {
				return
			}
			list.nextFetchTimes = append(list.nextFetchTimes, nextFetchTime.UTC())
			list.interruptChans = append(list.interruptChans, make(chan struct{}))
			list.updatedAt = append(list.updatedAt, lastUpdated)
			list.statuses = append(list.statuses, ComicNotUpdating)
		})
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	transaction.Commit()

	list.ScheduleComicFetches() //TODO: only the new ones
	return
}
