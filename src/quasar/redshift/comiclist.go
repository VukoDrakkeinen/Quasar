package redshift

import (
	"database/sql"
	"log"
	"quasar/qutils"
	"quasar/redshift/idbase"
	"quasar/redshift/qdb"
	"strings"
	"time"
)

type ComicList []Comic

func (this ComicList) createDB(db *sql.DB) {
	createCmd := `
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
		image TEXT,
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
	_, err := db.Exec(createCmd)
	if err != nil {
		log.Println(err)
	}
}

func (this ComicList) SaveToDB() { //TODO: update db entries instead of duplicating them
	log.SetFlags(log.Lshortfile | log.Ltime)
	db := qdb.DB() //TODO: error out on nil
	this.createDB(db)
	db.Exec("PRAGMA foreign_keys = ON;")

	lastIdCmd := `SELECT last_insert_rowid();`

	idsInsertionCmd := `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	type tuple struct {
		dict      qdb.InsertionStmtExecutor
		tableName string
		colName   string
	}
	for _, tuple := range []tuple{
		{&idbase.LangDict, "langs", "lang"}, //TODO?: global state, hmm
		{&idbase.Scanlators, "scanlators", "scanlator"},
		{&idbase.Authors, "authors", "author"},
		{&idbase.Artists, "artists", "artist"},
		{&idbase.ComicGenres, "genres", "genre"},
		{&idbase.ComicTags, "tags", "tag"},
	} {
		transaction, _ := db.Begin()
		rep := strings.NewReplacer("$tableName", tuple.tableName, "$colName", tuple.colName)
		idsInsertionStmt, _ := transaction.Prepare(rep.Replace(idsInsertionCmd))
		tuple.dict.ExecuteInsertionStmt(idsInsertionStmt)
		idsInsertionStmt.Close()
		transaction.Commit()
	}

	comicsInsertionCmd := `
	INSERT INTO comics(
		title, type, status, scanStatus, desc, rating, mature, image,
		useDefaultsBits, notifMode, accumCount, delayDuration
	) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	altTitlesInsertionCmd := `INSERT INTO altTitles(title) VALUES(?);`
	altTitlesRelationCmd := `INSERT INTO rel_Comic_AltTitles(comicId, titleId) VALUES(?, ?);`
	authorsRelationCmd := `INSERT INTO rel_Comic_Authors(comicId, authorId) VALUES(?, ?);`
	artistsRelationCmd := `INSERT INTO rel_Comic_Artists(comicId, artistId) VALUES(?, ?);`
	genresRelationCmd := `INSERT INTO rel_Comic_Genres(comicId, genreId) VALUES(?, ?);`
	tagsRelationCmd := `INSERT INTO rel_Comic_Tags(comicId, tagId) VALUES(?, ?);`
	sourcesInsertionCmd := `INSERT INTO sources(pluginName, url, markAsRead) VALUES(?, ?, ?);`
	sourcesRelationCmd := `INSERT INTO rel_Comic_Sources(comicId, sourceId) VALUES(?, ?);`
	chaptersInsertionCmd := `INSERT INTO chapters(identity, alreadyRead) VALUES(?, ?);`
	chaptersRelationCmd := `INSERT INTO rel_Comic_Chapters(comicId, chapterId) VALUES(?, ?);`
	scanlationInsertionCmd := `INSERT INTO scanlations(title, lang, pluginName, url) VALUES(?, ?, ?, ?);`
	scanlationRelationCmd := `INSERT INTO rel_Chapter_Scanlations(chapterId, scanlationid) VALUES(?, ?);`
	scanlatorsRelationCmd := `INSERT INTO rel_Scanlation_Scanlators(scanlationId, scanlatorId) VALUES(?, ?);`
	pageLinksInsertionCmd := `INSERT INTO pageLinks(url) VALUES(?);`
	pageLinksRelationCmd := `INSERT INTO rel_Scanlation_PageLinks(scanlationId, pageLinkId) VALUES(?, ?);`
	for _, comic := range this { //TODO: do not reprepare statements every iteration. Make them transaction-specific instead.
		transaction, _ := db.Begin()

		lastIdStmt, _ := transaction.Prepare(lastIdCmd)

		var comicId int
		comicInsertionStmt, _ := transaction.Prepare(comicsInsertionCmd)
		stts := &comic.Settings
		inf := &comic.Info
		if !stts.Valid() {
			stts = NewIndividualSettings(LoadGlobalSettings()) //FIXME: use current globals, not saved!
		}
		comicInsertionStmt.Exec(
			inf.Title, inf.Type, inf.Status, inf.ScanlationStatus, inf.Description, //Info
			inf.Rating, inf.Mature, "TODO", //FIXME: get path to thumbnail

			qutils.BoolsToBitfield(stts.UseDefaults), stts.UpdateNotificationMode, //Settings
			stts.AccumulativeModeCount, stts.DelayedModeDuration,
		)
		lastIdStmt.QueryRow().Scan(&comicId)
		comicInsertionStmt.Close()

		altTitlesInsertionStmt, _ := transaction.Prepare(altTitlesInsertionCmd)
		altTitlesRelationStmt, _ := transaction.Prepare(altTitlesRelationCmd)
		for title, _ := range inf.AltTitles {
			altTitlesInsertionStmt.Exec(title)
			var titleId int
			lastIdStmt.QueryRow().Scan(&titleId)
			altTitlesRelationStmt.Exec(comicId, titleId)
		}
		altTitlesInsertionStmt.Close()
		altTitlesRelationStmt.Close()

		authorsRelationStmt, _ := transaction.Prepare(authorsRelationCmd)
		for _, authorId := range inf.Authors {
			authorId.ExecuteInsertionStmt(authorsRelationStmt, comicId)
		}
		authorsRelationStmt.Close()

		artistsRelationStmt, _ := transaction.Prepare(artistsRelationCmd)
		for _, artistId := range inf.Artists {
			artistId.ExecuteInsertionStmt(artistsRelationStmt, comicId)
		}
		artistsRelationStmt.Close()

		genresRelationStmt, _ := transaction.Prepare(genresRelationCmd)
		for genreId := range inf.Genres {
			genreId.ExecuteInsertionStmt(genresRelationStmt, comicId)
		}
		genresRelationStmt.Close()

		tagsRelationStmt, _ := transaction.Prepare(tagsRelationCmd)
		for tagId := range inf.Categories {
			tagId.ExecuteInsertionStmt(tagsRelationStmt, comicId)
		}
		tagsRelationStmt.Close()

		sourcesInsertionStmt, _ := transaction.Prepare(sourcesInsertionCmd)
		sourcesRelationStmt, _ := transaction.Prepare(sourcesRelationCmd)
		for _, src := range comic.sources {
			sourcesInsertionStmt.Exec(string(src.PluginName), src.URL, src.MarkAsRead)
			var sourceId int
			lastIdStmt.QueryRow().Scan(&sourceId)
			sourcesRelationStmt.Exec(comicId, sourceId)
		}
		sourcesInsertionStmt.Close()
		sourcesRelationStmt.Close()

		chaptersInsertionStmt, _ := transaction.Prepare(chaptersInsertionCmd)
		chaptersRelationStmt, _ := transaction.Prepare(chaptersRelationCmd)
		scanlationInsertionStmt, _ := transaction.Prepare(scanlationInsertionCmd)
		scanlationRelationStmt, _ := transaction.Prepare(scanlationRelationCmd)
		scanlatorsRelationStmt, _ := transaction.Prepare(scanlatorsRelationCmd)
		pageLinksInsertionStmt, _ := transaction.Prepare(pageLinksInsertionCmd)
		pageLinksRelationStmt, _ := transaction.Prepare(pageLinksRelationCmd)
		for i := 0; i < comic.ChapterCount(); i++ {
			chapter, identity := comic.GetChapter(i)
			chaptersInsertionStmt.Exec(identity.n(), chapter.AlreadyRead)
			var chapterId int
			lastIdStmt.QueryRow().Scan(&chapterId)
			chaptersRelationStmt.Exec(comicId, chapterId)

			for i := 0; i < chapter.ScanlationsCount(); i++ {
				sc := chapter.Scanlation(i)
				sc.Language.ExecuteInsertionStmt(scanlationInsertionStmt, sc.Title, string(sc.PluginName), sc.URL)
				var scanlationId int
				lastIdStmt.QueryRow().Scan(&scanlationId)
				scanlationRelationStmt.Exec(chapterId, scanlationId)

				for _, scanlator := range sc.Scanlators.ToSlice() {
					scanlator.ExecuteInsertionStmt(scanlatorsRelationStmt, scanlationId)
				}

				for _, pageLink := range sc.PageLinks {
					pageLinksInsertionStmt.Exec(pageLink)
					var pageLinkId int
					lastIdStmt.QueryRow().Scan(&pageLinkId)
					pageLinksRelationStmt.Exec(scanlationId, pageLinkId)
				}
			}
		}
		chaptersInsertionStmt.Close()
		chaptersRelationStmt.Close()
		scanlationInsertionStmt.Close()
		scanlationRelationStmt.Close()
		scanlatorsRelationStmt.Close()
		pageLinksInsertionStmt.Close()
		pageLinksRelationStmt.Close()

		lastIdStmt.Close()

		transaction.Commit()
	}
}

func LoadComicList() (list ComicList, err error) {
	log.SetFlags(log.Ltime | log.Lshortfile)
	db := qdb.DB()
	list.createDB(db)

	transaction, _ := db.Begin()

	idsQueryCmd := `SELECT $colName FROM $tableName;`
	type tuple struct {
		dict      qdb.QueryStmtExecutor
		tableName string
		colName   string
	}
	for _, tuple := range []tuple{ //TODO?: dicts as function arguments? (global state side effects are not nice)
		{&idbase.LangDict, "langs", "lang"},
		{&idbase.Scanlators, "scanlators", "scanlator"},
		{&idbase.Authors, "authors", "author"},
		{&idbase.Artists, "artists", "artist"},
		{&idbase.ComicGenres, "genres", "genre"},
		{&idbase.ComicTags, "tags", "tag"},
	} {
		rep := strings.NewReplacer("$tableName", tuple.tableName, "$colName", tuple.colName)
		idsQueryStmt, _ := transaction.Prepare(rep.Replace(idsQueryCmd))
		tuple.dict.ExecuteQueryStmt(idsQueryStmt)
		idsQueryStmt.Close()
	}

	comicsQueryCmd := `
	SELECT
		id,
		title, type, status, scanStatus, desc, rating, mature, image,
		useDefaultsBits, notifMode, accumCount, delayDuration
	FROM comics;`
	altTitlesQueryCmd := `SELECT title FROM altTitles WHERE id IN (SELECT titleId FROM rel_Comic_AltTitles WHERE comicId = ?);`
	authorsQueryCmd := `SELECT authorId FROM rel_Comic_Authors WHERE comicId = ?;`
	artistsQueryCmd := `SELECT artistId FROM rel_Comic_Artists WHERE comicId = ?;`
	genresQueryCmd := `SELECT genreId FROM rel_Comic_Genres WHERE comicId = ?;`
	tagsQueryCmd := `SELECT tagId FROM rel_Comic_Tags WHERE comicId = ?;`
	sourcesQueryCmd := `
	SELECT
		pluginName, url, markAsRead
	FROM sources
	WHERE id IN(
		SELECT sourceId FROM rel_Comic_Sources WHERE comicId = ?
	);`
	chaptersQueryCmd := `
	SELECT
		id, identity, alreadyRead
	FROM chapters
	WHERE id IN(
		SELECT chapterId FROM rel_Comic_Chapters WHERE comicId = ?
	);`
	scanlationsQueryCmd := `
	SELECT
		id, title, lang, pluginName, url
	FROM scanlations
	WHERE id IN(
		SELECT scanlationId FROM rel_Chapter_Scanlations WHERE chapterId = ?
	);`
	scanlatorsQueryCmd := `SELECT scanlatorId FROM rel_Scanlation_Scanlators WHERE scanlationId = ?;`
	pageLinksQueryCmd := `
	SELECT url
	FROM pageLinks
	WHERE id IN(
		SELECT pageLinkId FROM rel_Scanlation_PageLinks WHERE scanlationId = ?
	);`

	comicsQueryStmt, _ := transaction.Prepare(comicsQueryCmd)

	comicRows, _ := comicsQueryStmt.Query()
	for comicRows.Next() {
		info := ComicInfo{}
		stts := IndividualSettings{}
		var imagePath sql.NullString //FIXME
		var bitfield uint64
		var comicId int
		var duration int64
		comicRows.Scan(
			&comicId,
			&info.Title, &info.Type, &info.Status, &info.ScanlationStatus, &info.Description, &info.Rating, &info.Mature, &imagePath,
			&bitfield, &stts.UpdateNotificationMode, &stts.AccumulativeModeCount, &duration,
		)
		stts.DelayedModeDuration = time.Duration(duration)
		stts.UseDefaults = qutils.BitfieldToBools(bitfield)

		altTitlesQueryStmt, _ := transaction.Prepare(altTitlesQueryCmd)
		altTitleRows, _ := altTitlesQueryStmt.Query(comicId)
		altTitles := make(map[string]struct{})
		for altTitleRows.Next() {
			var title string
			altTitleRows.Scan(&title)
			altTitles[title] = struct{}{}
		}
		info.AltTitles = altTitles

		authorsQueryStmt, _ := transaction.Prepare(authorsQueryCmd)
		authorRows, _ := authorsQueryStmt.Query(comicId)
		var authors []idbase.AuthorId
		for authorRows.Next() {
			var authorId idbase.AuthorId
			authorRows.Scan(&authorId)
			authors = append(authors, authorId)
		}
		info.Authors = authors

		artistsQueryStmt, _ := transaction.Prepare(artistsQueryCmd)
		artistRows, _ := artistsQueryStmt.Query(comicId)
		var artists []idbase.ArtistId
		for artistRows.Next() {
			var artistId idbase.ArtistId
			artistRows.Scan(&artistId)
			artists = append(artists, artistId)
		}
		info.Artists = artists

		genresQueryStmt, _ := transaction.Prepare(genresQueryCmd)
		genreRows, _ := genresQueryStmt.Query(comicId)
		genres := make(map[idbase.ComicGenreId]struct{})
		for genreRows.Next() {
			var genre idbase.ComicGenreId
			genreRows.Scan(&genre)
			genres[genre] = struct{}{}
		}
		info.Genres = genres

		tagsQueryStmt, _ := transaction.Prepare(tagsQueryCmd)
		tagRows, _ := tagsQueryStmt.Query(comicId)
		tags := make(map[idbase.ComicTagId]struct{})
		for tagRows.Next() {
			var tag idbase.ComicTagId
			tagRows.Scan(&tag)
			tags[tag] = struct{}{}
		}
		info.Categories = tags

		comic := Comic{
			Info:     info,
			Settings: stts,
		}

		sourcesQueryStmt, _ := transaction.Prepare(sourcesQueryCmd)
		sourceRows, _ := sourcesQueryStmt.Query(comicId)
		for sourceRows.Next() {
			var source UpdateSource
			sourceRows.Scan(&source.PluginName, &source.URL, &source.MarkAsRead)
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
			var chapterId int
			var identity ChapterIdentity
			chapter := Chapter{}
			chapterRows.Scan(&chapterId, &identity, &chapter.AlreadyRead)

			scanlationRows, _ := scanlationsQueryStmt.Query(chapterId)
			for scanlationRows.Next() {
				var scanlationId int
				scanlation := ChapterScanlation{}
				scanlationRows.Scan(&scanlationId, &scanlation.Title, &scanlation.Language, &scanlation.PluginName, &scanlation.URL)

				scanlatorRows, _ := scanlatorsQueryStmt.Query(scanlationId)
				var scanlators []idbase.ScanlatorId
				for scanlatorRows.Next() {
					var scanlator idbase.ScanlatorId
					scanlatorRows.Scan(&scanlator)
					scanlators = append(scanlators, scanlator)
				}
				scanlation.Scanlators = idbase.JoinScanlators(scanlators)

				pageLinkRows, _ := pageLinksQueryStmt.Query(scanlationId)
				for pageLinkRows.Next() {
					var pageLink string
					pageLinkRows.Scan(&pageLink)
					scanlation.PageLinks = append(scanlation.PageLinks, pageLink)
				}

				chapter.AddScanlation(scanlation)
			}
			identities = append(identities, identity)
			chapters = append(chapters, chapter)
		}
		comic.AddMultipleChapters(identities, chapters)

		list = append(list, comic)
	}

	transaction.Commit()
	return //FIXME: return errors
}
