package redshift

import (
	"database/sql"
	"fmt"
	"log"
	"quasar/qutils"
	"quasar/redshift/idsdict"
	"quasar/redshift/qdb"
	"strings"
	"time"
)

var (
	createCmd = `
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
	);
	CREATE TABLE IF NOT EXISTS chapters(
		id INTEGER PRIMARY KEY,
		identity INTEGER UNIQUE NOT NULL,
		alreadyRead INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Chapters(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		chapterId INTEGER NOT NULL UNIQUE REFERENCES chapters(id) ON DELETE CASCADE,
		CONSTRAINT pk_CO_CH PRIMARY KEY (comicId, chapterId)
	);
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

	enableForeignKeysCmd = `PRAGMA foreign_keys = ON;`

	idsInsertionPreCmd = `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	comicsInsertionCmd = `
	INSERT OR REPLACE INTO comics(
		id,
		title, type, status, scanStatus, desc, rating, mature, thumbnailFilename,
		useDefaultsBits, notifMode, accumCount, delayDuration
	) VALUES((SELECT id FROM comics WHERE id = ?), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	altTitlesInsertionCmd  = `INSERT OR REPLACE INTO altTitles(id, title) VALUES((SELECT id FROM altTitles WHERE id = ?), ?);`
	altTitlesRelationCmd   = `INSERT OR IGNORE INTO rel_Comic_AltTitles(comicId, titleId) VALUES(?, ?);`
	authorsRelationCmd     = `INSERT OR IGNORE INTO rel_Comic_Authors(comicId, authorId) VALUES(?, ?);`
	artistsRelationCmd     = `INSERT OR IGNORE INTO rel_Comic_Artists(comicId, artistId) VALUES(?, ?);`
	genresRelationCmd      = `INSERT OR IGNORE INTO rel_Comic_Genres(comicId, genreId) VALUES(?, ?);`
	tagsRelationCmd        = `INSERT OR IGNORE INTO rel_Comic_Tags(comicId, tagId) VALUES(?, ?);`
	sourcesInsertionCmd    = `INSERT OR REPLACE INTO sources(id, pluginName, url, markAsRead) VALUES((SELECT id FROM sources WHERE id = ?), ?, ?, ?);`
	sourcesRelationCmd     = `INSERT OR IGNORE INTO rel_Comic_Sources(comicId, sourceId) VALUES(?, ?);`
	chaptersInsertionCmd   = `INSERT OR REPLACE INTO chapters(id, identity, alreadyRead) VALUES((SELECT id FROM chapters WHERE id = ?), ?, ?);`
	chaptersRelationCmd    = `INSERT OR IGNORE INTO rel_Comic_Chapters(comicId, chapterId) VALUES(?, ?);`
	scanlationInsertionCmd = `INSERT OR REPLACE INTO scanlations(id, title, lang, pluginName, url) VALUES((SELECT id FROM scanlations WHERE id = ?), ?, ?, ?, ?);`
	scanlationRelationCmd  = `INSERT OR IGNORE INTO rel_Chapter_Scanlations(chapterId, scanlationid) VALUES(?, ?);`
	scanlatorsRelationCmd  = `INSERT OR IGNORE INTO rel_Scanlation_Scanlators(scanlationId, scanlatorId) VALUES(?, ?);`
	pageLinksInsertionCmd  = `INSERT OR REPLACE INTO pageLinks(id, url) VALUES((SELECT id FROM pageLinks WHERE id = ?), ?);`
	pageLinksRelationCmd   = `INSERT OR IGNORE INTO rel_Scanlation_PageLinks(scanlationId, pageLinkId) VALUES(?, ?);`

	idsQueryPreCmd = `SELECT $colName FROM $tableName;` //TODO?: use placeholders?
	comicsQueryCmd = `
	SELECT
		id,
		title, type, status, scanStatus, desc, rating, mature, thumbnailFilename,
		useDefaultsBits, notifMode, accumCount, delayDuration
	FROM comics;`
	altTitlesQueryCmd = `SELECT id, title FROM altTitles WHERE id IN (SELECT titleId FROM rel_Comic_AltTitles WHERE comicId = ?);`
	authorsQueryCmd   = `SELECT authorId FROM rel_Comic_Authors WHERE comicId = ?;`
	artistsQueryCmd   = `SELECT artistId FROM rel_Comic_Artists WHERE comicId = ?;`
	genresQueryCmd    = `SELECT genreId FROM rel_Comic_Genres WHERE comicId = ?;`
	tagsQueryCmd      = `SELECT tagId FROM rel_Comic_Tags WHERE comicId = ?;`
	sourcesQueryCmd   = `
	SELECT
		id, pluginName, url, markAsRead
	FROM sources
	WHERE id IN(
		SELECT sourceId FROM rel_Comic_Sources WHERE comicId = ?
	);`
	chaptersQueryCmd = `
	SELECT
		id, identity, alreadyRead
	FROM chapters
	WHERE id IN(
		SELECT chapterId FROM rel_Comic_Chapters WHERE comicId = ?
	);`
	scanlationsQueryCmd = `
	SELECT
		id, title, lang, pluginName, url
	FROM scanlations
	WHERE id IN(
		SELECT scanlationId FROM rel_Chapter_Scanlations WHERE chapterId = ?
	);`
	scanlatorsQueryCmd = `SELECT scanlatorId FROM rel_Scanlation_Scanlators WHERE scanlationId = ?;`
	pageLinksQueryCmd  = `
	SELECT id, url
	FROM pageLinks
	WHERE id IN(
		SELECT pageLinkId FROM rel_Scanlation_PageLinks WHERE scanlationId = ?
	);`
)

type ComicList []*Comic

func (this ComicList) createDB(db *sql.DB) {
	transaction, _ := db.Begin()
	transaction.Exec(createCmd)
	transaction.Commit()
}

func (this ComicList) SaveToDB() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	db := qdb.DB() //TODO: error out on nil
	this.createDB(db)
	db.Exec(enableForeignKeysCmd)

	type tuple struct {
		dict      qdb.InsertionStmtExecutor
		tableName string
		colName   string
	}
	for _, tuple := range []tuple{
		{&idsdict.Langs, "langs", "lang"}, //TODO?: global state, hmm
		{&idsdict.Scanlators, "scanlators", "scanlator"},
		{&idsdict.Authors, "authors", "author"},
		{&idsdict.Artists, "artists", "artist"},
		{&idsdict.ComicGenres, "genres", "genre"},
		{&idsdict.ComicTags, "tags", "tag"},
	} {
		transaction, _ := db.Begin()
		rep := strings.NewReplacer("$tableName", tuple.tableName, "$colName", tuple.colName)
		idsInsertionStmt, _ := transaction.Prepare(rep.Replace(idsInsertionPreCmd))
		tuple.dict.ExecuteInsertionStmt(idsInsertionStmt)
		idsInsertionStmt.Close()
		transaction.Commit()
	}

	for _, comic := range this { //TODO: do not reprepare statements every iteration. Make them transaction-specific instead.
		transaction, _ := db.Begin()

		comicInsertionStmt, _ := transaction.Prepare(comicsInsertionCmd)
		altTitlesInsertionStmt, _ := transaction.Prepare(altTitlesInsertionCmd)
		altTitlesRelationStmt, _ := transaction.Prepare(altTitlesRelationCmd)
		authorsRelationStmt, _ := transaction.Prepare(authorsRelationCmd)
		artistsRelationStmt, _ := transaction.Prepare(artistsRelationCmd)
		genresRelationStmt, _ := transaction.Prepare(genresRelationCmd)
		tagsRelationStmt, _ := transaction.Prepare(tagsRelationCmd)
		sourcesInsertionStmt, _ := transaction.Prepare(sourcesInsertionCmd)
		sourcesRelationStmt, _ := transaction.Prepare(sourcesRelationCmd)
		chaptersInsertionStmt, _ := transaction.Prepare(chaptersInsertionCmd)
		chaptersRelationStmt, _ := transaction.Prepare(chaptersRelationCmd)
		scanlationInsertionStmt, _ := transaction.Prepare(scanlationInsertionCmd)
		scanlationRelationStmt, _ := transaction.Prepare(scanlationRelationCmd)
		scanlatorsRelationStmt, _ := transaction.Prepare(scanlatorsRelationCmd)
		pageLinksInsertionStmt, _ := transaction.Prepare(pageLinksInsertionCmd)
		pageLinksRelationStmt, _ := transaction.Prepare(pageLinksRelationCmd)

		stmts := InsertionStmtGroup{
			comicInsertionStmt:      comicInsertionStmt,
			altTitlesInsertionStmt:  altTitlesInsertionStmt,
			altTitlesRelationStmt:   altTitlesRelationStmt,
			authorsRelationStmt:     authorsRelationStmt,
			artistsRelationStmt:     artistsRelationStmt,
			genresRelationStmt:      genresRelationStmt,
			tagsRelationStmt:        tagsRelationStmt,
			sourcesInsertionStmt:    sourcesInsertionStmt,
			sourcesRelationStmt:     sourcesRelationStmt,
			chaptersInsertionStmt:   chaptersInsertionStmt,
			chaptersRelationStmt:    chaptersRelationStmt,
			scanlationInsertionStmt: scanlationInsertionStmt,
			scanlationRelationStmt:  scanlationRelationStmt,
			scanlatorsRelationStmt:  scanlatorsRelationStmt,
			pageLinksInsertionStmt:  pageLinksInsertionStmt,
			pageLinksRelationStmt:   pageLinksRelationStmt,
		}
		err := comic.SQLInsert(stmts)
		if err != nil {
			fmt.Println(err)
			transaction.Rollback()
		} else {
			transaction.Commit()
		}

		comicInsertionStmt.Close()
		altTitlesInsertionStmt.Close()
		altTitlesRelationStmt.Close()
		authorsRelationStmt.Close()
		artistsRelationStmt.Close()
		genresRelationStmt.Close()
		tagsRelationStmt.Close()
		sourcesInsertionStmt.Close()
		sourcesRelationStmt.Close()
		chaptersInsertionStmt.Close()
		chaptersRelationStmt.Close()
		scanlationInsertionStmt.Close()
		scanlationRelationStmt.Close()
		scanlatorsRelationStmt.Close()
		pageLinksInsertionStmt.Close()
		pageLinksRelationStmt.Close()
	}
}

func LoadComicList() (list ComicList, err error) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	db := qdb.DB()
	list.createDB(db)
	transaction, _ := db.Begin()

	type tuple struct {
		dict      qdb.QueryStmtExecutor
		tableName string
		colName   string
	}
	for _, tuple := range []tuple{ //TODO?: dicts as function arguments? (global state side effects are not nice)
		{&idsdict.Langs, "langs", "lang"},
		{&idsdict.Scanlators, "scanlators", "scanlator"},
		{&idsdict.Authors, "authors", "author"},
		{&idsdict.Artists, "artists", "artist"},
		{&idsdict.ComicGenres, "genres", "genre"},
		{&idsdict.ComicTags, "tags", "tag"},
	} {
		rep := strings.NewReplacer("$tableName", tuple.tableName, "$colName", tuple.colName)
		idsQueryStmt, _ := transaction.Prepare(rep.Replace(idsQueryPreCmd))
		tuple.dict.ExecuteQueryStmt(idsQueryStmt)
		idsQueryStmt.Close()
	}

	comicsQueryStmt, _ := transaction.Prepare(comicsQueryCmd)
	comicRows, _ := comicsQueryStmt.Query()
	for comicRows.Next() {
		info := ComicInfo{altSQLIds: make(map[string]int64)}
		stts := IndividualSettings{}
		var comicId int64
		var thumbnailFilename sql.NullString
		var useDefaultsBitfield uint64
		var duration int64
		comicRows.Scan(
			&comicId,
			&info.Title, &info.Type, &info.Status, &info.ScanlationStatus, &info.Description, &info.Rating, &info.Mature, &thumbnailFilename,
			&useDefaultsBitfield, &stts.UpdateNotificationMode, &stts.AccumulativeModeCount, &duration,
		)
		info.ThumbnailFilename = thumbnailFilename.String
		stts.DelayedModeDuration = time.Duration(duration)
		stts.UseDefaults = qutils.BitfieldToBools(useDefaultsBitfield)

		altTitlesQueryStmt, _ := transaction.Prepare(altTitlesQueryCmd)
		altTitleRows, _ := altTitlesQueryStmt.Query(comicId)
		altTitles := make(map[string]struct{})
		for altTitleRows.Next() {
			var titleId int64
			var title string
			altTitleRows.Scan(&titleId, &title)
			altTitles[title] = struct{}{}
			info.altSQLIds[title] = titleId
		}
		info.AltTitles = altTitles

		authorsQueryStmt, _ := transaction.Prepare(authorsQueryCmd)
		authorRows, _ := authorsQueryStmt.Query(comicId)
		var authors []idsdict.AuthorId
		for authorRows.Next() {
			var author idsdict.AuthorId
			authorRows.Scan(&author)
			authors = append(authors, author)
		}
		info.Authors = authors

		artistsQueryStmt, _ := transaction.Prepare(artistsQueryCmd)
		artistRows, _ := artistsQueryStmt.Query(comicId)
		var artists []idsdict.ArtistId
		for artistRows.Next() {
			var artist idsdict.ArtistId
			artistRows.Scan(&artist)
			artists = append(artists, artist)
		}
		info.Artists = artists

		genresQueryStmt, _ := transaction.Prepare(genresQueryCmd)
		genreRows, _ := genresQueryStmt.Query(comicId)
		genres := make(map[idsdict.ComicGenreId]struct{})
		for genreRows.Next() {
			var genre idsdict.ComicGenreId
			genreRows.Scan(&genre)
			genres[genre] = struct{}{}
		}
		info.Genres = genres

		tagsQueryStmt, _ := transaction.Prepare(tagsQueryCmd)
		tagRows, _ := tagsQueryStmt.Query(comicId)
		tags := make(map[idsdict.ComicTagId]struct{})
		for tagRows.Next() {
			var tag idsdict.ComicTagId
			tagRows.Scan(&tag)
			tags[tag] = struct{}{}
		}
		info.Categories = tags

		comic := NewComic()
		comic.Info = info
		comic.Settings = stts
		comic.sqlId = comicId

		sourcesQueryStmt, _ := transaction.Prepare(sourcesQueryCmd)
		sourceRows, _ := sourcesQueryStmt.Query(comicId)
		for sourceRows.Next() {
			var sourceId int64
			var source UpdateSource
			sourceRows.Scan(&sourceId, &source.PluginName, &source.URL, &source.MarkAsRead)
			source.sqlId = sourceId
			comic.AddSource(source)
		}

		chaptersQueryStmt, _ := transaction.Prepare(chaptersQueryCmd)
		scanlationsQueryStmt, _ := transaction.Prepare(scanlationsQueryCmd)
		scanlatorsQueryStmt, _ := transaction.Prepare(scanlatorsQueryCmd)
		pageLinksQueryStmt, _ := transaction.Prepare(pageLinksQueryCmd)
		chapterRows, _ := chaptersQueryStmt.Query(comicId)
		var identities []ChapterIdentity
		var chapters []Chapter
		for chapterRows.Next() {
			var chapterId int64
			var identity ChapterIdentity
			chapter := NewChapter(false)
			chapterRows.Scan(&chapterId, &identity, &chapter.AlreadyRead)
			chapter.sqlId = chapterId

			scanlationRows, _ := scanlationsQueryStmt.Query(chapterId)
			for scanlationRows.Next() {
				var scanlationId int64
				scanlation := ChapterScanlation{}
				scanlationRows.Scan(&scanlationId, &scanlation.Title, &scanlation.Language, &scanlation.PluginName, &scanlation.URL)
				scanlation.sqlId = scanlationId

				scanlatorRows, _ := scanlatorsQueryStmt.Query(scanlationId)
				var scanlators []idsdict.ScanlatorId
				for scanlatorRows.Next() {
					var scanlator idsdict.ScanlatorId
					scanlatorRows.Scan(&scanlator)
					scanlators = append(scanlators, scanlator)
				}
				scanlation.Scanlators = idsdict.JoinScanlators(scanlators)

				pageLinkRows, _ := pageLinksQueryStmt.Query(scanlationId)
				for pageLinkRows.Next() {
					var pageLinkId int64
					var pageLink string
					pageLinkRows.Scan(&pageLinkId, &pageLink)
					scanlation.PageLinks = append(scanlation.PageLinks, pageLink)
					scanlation.plSQLIds = append(scanlation.plSQLIds, pageLinkId)
				}

				chapter.AddScanlation(scanlation)
			}

			identities = append(identities, identity)
			chapters = append(chapters, *chapter)
		}
		comic.AddMultipleChapters(identities, chapters)

		list = append(list, comic)
	}

	transaction.Commit()
	return //FIXME: return errors
}

type InsertionStmtGroup struct {
	lastIdStmt              *sql.Stmt
	comicInsertionStmt      *sql.Stmt
	altTitlesInsertionStmt  *sql.Stmt
	altTitlesRelationStmt   *sql.Stmt
	authorsRelationStmt     *sql.Stmt
	artistsRelationStmt     *sql.Stmt
	genresRelationStmt      *sql.Stmt
	tagsRelationStmt        *sql.Stmt
	sourcesInsertionStmt    *sql.Stmt
	sourcesRelationStmt     *sql.Stmt
	chaptersInsertionStmt   *sql.Stmt
	chaptersRelationStmt    *sql.Stmt
	scanlationInsertionStmt *sql.Stmt
	scanlationRelationStmt  *sql.Stmt
	scanlatorsRelationStmt  *sql.Stmt
	pageLinksInsertionStmt  *sql.Stmt
	pageLinksRelationStmt   *sql.Stmt
}

type QueryStmtGroup struct {
	comicsQueryStmt      *sql.Stmt
	altTitlesQueryStmt   *sql.Stmt
	authorsQueryStmt     *sql.Stmt
	artistsQueryStmt     *sql.Stmt
	genresQueryStmt      *sql.Stmt
	tagsQueryStmt        *sql.Stmt
	sourcesQueryStmt     *sql.Stmt
	chaptersQueryStmt    *sql.Stmt
	scanlationsQueryStmt *sql.Stmt
	scanlatorsQueryStmt  *sql.Stmt
	pageLinksQueryStmt   *sql.Stmt
}
