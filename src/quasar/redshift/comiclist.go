package redshift

import (
	"fmt"
	"math"
	"quasar/datadir/qdb"
	"quasar/datadir/qlog"
	"quasar/qutils/qerr"
	"quasar/redshift/idsdict"
	"strings"
	"time"
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
		nextFetchTime TIMESTAMP NOT NULL
	);
	`
	idsInsertionPreCmd   = `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	idsQueryPreCmd       = `SELECT $colName FROM $tableName;` //TODO?: use placeholders?
	scheduleInsertionCmd = `INSERT OR REPLACE INTO schedule(comicId, nextFetchTime) VALUES((SELECT id FROM comics WHERE id = ?), ?);`
	scheduleQueryCmd     = `SELECT nextFetchTime FROM schedule WHERE comicId = ?;`
)

func NewComicList(fetcher *fetcher) ComicList {
	return ComicList{
		comics:         make([]*Comic, 0, 10),
		interruptChans: make([]chan struct{}, 0, 10),
		fetcher:        fetcher,
	}
}

type ComicList struct {
	comics         []*Comic
	updatedAt      []time.Time
	nextFetchTimes []time.Time
	interruptChans []chan struct{}
	fetcher        *fetcher
}

func (this *ComicList) AddComics(comics []*Comic) {
	this.comics = append(this.comics, comics...)
	interruptChans := make([]chan struct{}, len(comics))
	for i := range interruptChans {
		interruptChans[i] = make(chan struct{})
	}
	this.interruptChans = append(this.interruptChans, interruptChans...)
	this.nextFetchTimes = append(this.nextFetchTimes, make([]time.Time, len(comics))...)
	this.updatedAt = append(this.updatedAt, make([]time.Time, len(comics))...)
}

func (this *ComicList) RemoveComics(index, count int64) {
	this.comics = append(this.comics[:index], this.comics[index+count:]...)
	this.interruptChans = append(this.interruptChans[:index], this.interruptChans[index+count:]...)
	this.nextFetchTimes = append(this.nextFetchTimes[:index], this.nextFetchTimes[index+count:]...)
	this.updatedAt = append(this.updatedAt[:index], this.updatedAt[index+count:]...)
}

func (this ComicList) ComicLastUpdated(idx int) time.Time {
	return this.updatedAt[idx]
}

//TODO: remove
func (this ComicList) Hack_Comics() []*Comic {
	return this.comics
}

func (this ComicList) ScheduleComicFetches() {
	this.cancelSchedule()

	for i, comic := range this.comics {
		cSettings := comic.Settings()
		fSettings := this.fetcher.settings
		var duration time.Duration
		if overrideFrequency := cSettings.OverrideDefaults[2]; overrideFrequency {
			duration = cSettings.FetchFrequency
		} else {
			duration = fSettings.FetchFrequency
		}
		fetchOnStartup := fSettings.FetchOnStartup
		intervalFetching := fSettings.IntervalFetching

		go func() {
			i := i
			fmt.Println("  Old schedule", this.nextFetchTimes[i])
			now := time.Now().UTC()
			if fetchOnStartup {
				fmt.Println("  Fetch On Startup: Enabled")
				this.fetcher.DownloadChapterListFor(comic)
				this.updatedAt[i] = now //TODO?: actual now?
				this.nextFetchTimes[i] = now.Add(duration)
			} else if this.nextFetchTimes[i].Before(now) {
				fmt.Println("  Scheduled time in the past; adjusting...")
				this.fetcher.DownloadChapterListFor(comic)
				this.updatedAt[i] = now
				multiplier := divThenCeil(now.Sub(this.nextFetchTimes[i]), duration)
				this.nextFetchTimes[i].Add(multiplier * duration)
			}
			fmt.Println("  Scheduled fetch for", this.nextFetchTimes[i])

			if intervalFetching {
				fmt.Println("#  Interval Fetching Task Started")
				for {
					select {
					case <-time.After(this.nextFetchTimes[i].Sub(now)):
						this.fetcher.DownloadChapterListFor(comic)
						now := time.Now().UTC()
						this.updatedAt[i] = now
						this.nextFetchTimes[i] = now.Add(duration)
					case <-this.interruptChans[i]:
						return
					}
				}
			}
		}()
	}
}

/*
func (this ComicList) RescheduleComicFetch(comicIdx int) {
	close(this.interruptChans[comicIdx])
	this.ScheduleComicFetches() //TODO: just one, jeez
}
//*/

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

func (this ComicList) SaveToDB() { //TODO: write a unit test
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
	fmt.Println("\tIds write complete.")
	fmt.Println("\t Writing", len(this.comics), "comics")

	scheduleInsertionStmt := db.MustPrepare(scheduleInsertionCmd)
	dbStmts := SQLComicInsertStmts(db)
	defer dbStmts.Close()
	for i, comic := range this.comics {
		transaction, _ := db.Begin()
		stmts := dbStmts.ToTransactionSpecific(transaction)

		err := comic.SQLInsert(stmts)
		if err != nil { // no need to manually close statements, Commit() or Rollback() take care of that
			fmt.Println("Error while saving, rolling back")
			transaction.Rollback()
			fmt.Println(qerr.NewLocated(err))
		} else {
			transaction.Commit()
			fmt.Println("  Saving scheduled fetch time", this.nextFetchTimes[i])
			_, err := scheduleInsertionStmt.Exec(comic.sqlId, this.nextFetchTimes[i])
			if err != nil {
				fmt.Println(qerr.NewLocated(err))
			}
		}

	}
}

func (list ComicList) LoadFromDB() (err error) {
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
		list.comics = append(list.comics, comic)
		var nextFetchTime time.Time
		err = scheduleQueryStmt.QueryRow(comic.sqlId).Scan(&nextFetchTime)
		nextFetchTime = nextFetchTime.UTC()
		list.nextFetchTimes = append(list.nextFetchTimes, nextFetchTime)
		fmt.Println("  Loaded scheduled fetch time", list.nextFetchTimes[len(list.nextFetchTimes)-1])
		if err != nil {
			return qerr.NewLocated(err)
		}
	}

	transaction.Commit()
	return
}

func (this *ComicList) cancelSchedule() {
	for i, interrupt := range this.interruptChans {
		close(interrupt)
		this.interruptChans[i] = make(chan struct{})
	}
}

func divThenCeil(divident, divisor time.Duration) (multiplier time.Duration) {
	x := float64(divident)
	y := float64(divisor)
	return time.Duration(math.Ceil(x / y))
}
