package redshift

import (
	"database/sql"
	"log"
	"quasar/qutils"
	"quasar/redshift/idbase"
	"quasar/redshift/qdb"
	"strings"
)

type ComicList []Comic

func (this ComicList) createDB(db *sql.DB) {
	createIdsCmd := `
	CREATE TABLE IF NOT EXISTS ids_$table(
		id INTEGER PRIMARY KEY,
		name TEXT UNIQUE NOT NULL
	);`
	for _, tableName := range []string{"langs", "scanlators"} {
		_, err := db.Exec(strings.Replace(createIdsCmd, "$table", tableName, 1))
		if err != nil {
			log.Println(err)
		}
	}

	createCmd := `
	CREATE TABLE IF NOT EXISTS comic_Settings(
		id INTEGER PRIMARY KEY,
		useDefaultsBits INTEGER NOT NULL,
		notifMode INTEGER,
		accumCount INTEGER,
		delayDuration INTEGER,
		downloadsPath TEXT
	);
	CREATE TABLE IF NOT EXISTS comic_Infos(
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		type INTEGER NOT NULL,
		status INTEGER NOT NULL,
		scanStatus INTEGER NOT NULL,
		desc TEXT NOT NULL,
		rating REAL NOT NULL,
		mature INTEGER NOT NULL,
		image TEXT
	);
	CREATE TABLE IF NOT EXISTS info_AltTitles(
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS info_Authors(
		id INTEGER PRIMARY KEY,
		author TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS info_Artists(
		id INTEGER PRIMARY KEY,
		artist TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS info_Genres(
		id INTEGER PRIMARY KEY,
		genre TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS info_Tags(
		id INTEGER PRIMARY KEY,
		tag TEXT UNIQUE NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Info_AltTitles(
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id) ON DELETE CASCADE,
		titleId INTEGER NOT NULL REFERENCES info_AltTitles(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AT PRIMARY KEY (infoId, titleId)
	);
	CREATE TABLE IF NOT EXISTS rel_Info_Authors(
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id) ON DELETE CASCADE,
		authorId INTEGER NOT NULL REFERENCES info_Authors(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AU PRIMARY KEY (infoId, authorId)
	);
	CREATE TABLE IF NOT EXISTS rel_Info_Artists(
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id) ON DELETE CASCADE,
		artistId INTEGER NOT NULL REFERENCES info_Artists(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AR PRIMARY KEY (infoId, artistId)
	);
	CREATE TABLE IF NOT EXISTS rel_Info_Genres(
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id) ON DELETE CASCADE,
		genreId INTEGER NOT NULL REFERENCES info_Genres(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_GE PRIMARY KEY (infoId, genreId)
	);
	CREATE TABLE IF NOT EXISTS rel_Info_Tags(
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id) ON DELETE CASCADE,
		tagId INTEGER NOT NULL REFERENCES info_Tags(id) ON DELETE CASCADE,
		CONSTRAINT pk_CI_AT PRIMARY KEY (infoId, tagId)
	);
	CREATE TABLE IF NOT EXISTS comics(
		id INTEGER PRIMARY KEY,
		infoId INTEGER NOT NULL REFERENCES comic_Infos(id),
		settingsId INTEGER NOT NULL REFERENCES comic_Settings(id)
	);
	CREATE TABLE IF NOT EXISTS comic_Sources(
		id INTEGER PRIMARY KEY,
		pluginName TEXT NOT NULL,
		url TEXT NOT NULL,
		markAsRead INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Sources(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		sourceId INTEGER NOT NULL REFERENCES comic_Sources(id) ON DELETE CASCADE,
		CONSTRAINT pk_CO_SO PRIMARY KEY (comicId, sourceId)
	);
	CREATE TABLE IF NOT EXISTS comic_Chapters(
		id INTEGER PRIMARY KEY,
		identity INTEGER UNIQUE NOT NULL,
		alreadyRead INTEGER NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Comic_Chapters(
		comicId INTEGER NOT NULL REFERENCES comics(id) ON DELETE CASCADE,
		chapterId INTEGER NOT NULL UNIQUE REFERENCES comic_Chapters(id) ON DELETE CASCADE,
		CONSTRAINT pk_CO_CH PRIMARY KEY (comicId, chapterId)
	);
	CREATE TABLE IF NOT EXISTS chapter_Scanlations(
		id INTEGER PRIMARY KEY,
		title TEXT NOT NULL,
		lang INTEGER NOT NULL DEFAULT 1 REFERENCES ids_langs(id) ON DELETE SET DEFAULT,
		pluginName TEXT NOT NULL,
		url TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Chapter_Scanlations(
		chapterId INTEGER NOT NULL REFERENCES comic_Chapters(id) ON DELETE CASCADE,
		scanlationId INTEGER NOT NULL REFERENCES chapter_Scanlations(id) ON DELETE CASCADE,
		CONSTRAINT pk_CH_SC PRIMARY KEY (chapterId, scanlationId)
	);
	CREATE TABLE IF NOT EXISTS rel_Scanlations_Scanlators(
		scanlationId INTEGER NOT NULL REFERENCES chapter_Scanlations(id) ON DELETE CASCADE,
		scanlatorId INTEGER NOT NULL REFERENCES ids_scanlators(id) ON DELETE CASCADE,
		CONSTRAINT pk_SC_SC PRIMARY KEY (scanlationId, scanlatorId)
	);
	CREATE TABLE IF NOT EXISTS scanlation_PageLinks(
		id INTEGER PRIMARY KEY,
		url TEXT NOT NULL
	);
	CREATE TABLE IF NOT EXISTS rel_Scanlations_PageLinks(
		scanlationId INTEGER NOT NULL REFERENCES chapter_Scanlations(id) ON DELETE CASCADE,
		pageLinkId INTEGER NOT NULL REFERENCES scanlation_PageLinks(id) ON DELETE CASCADE,
		CONSTRAINT pk_SC_PA PRIMARY KEY (scanlationId, pageLinkId)
	);`
	_, err := db.Exec(createCmd)
	if err != nil {
		log.Println(err)
	}
}

func (this ComicList) SaveToDB() {
	log.SetFlags(log.Lshortfile | log.Ltime)
	db := qdb.DB() //TODO: error out on nil
	this.createDB(db)
	db.Exec("PRAGMA foreign_keys = ON;")

	lastIdCmd := `SELECT last_insert_rowid();`

	idsInsertionCmd := `INSERT OR IGNORE INTO $tableName($colName) VALUES(?);`
	type tuple struct {
		dict      qdb.DBStatementExecutor
		tableName string
		colName   string
	}
	for _, tuple := range []tuple{
		{&idbase.LangDict, "ids_langs", "name"},
		{&idbase.Scanlators, "ids_scanlators", "name"},
		{&idbase.Authors, "info_Authors", "author"},
		{&idbase.Artists, "info_Artists", "artist"},
		{&idbase.ComicGenres, "info_Genres", "genre"},
		{&idbase.ComicTags, "info_Tags", "tag"},
	} {
		transaction, err := db.Begin()
		if err != nil {
			log.Println(err) //TODO: log error
		}
		rep := strings.NewReplacer("$tableName", tuple.tableName, "$colName", tuple.colName)
		idsInsertionStmt, err := transaction.Prepare(rep.Replace(idsInsertionCmd))
		if err != nil {
			log.Println(err) //TODO: log error
		}
		err = tuple.dict.ExecuteDBStatement(idsInsertionStmt)
		if err != nil {
			log.Println(err) //TODO: log error
		}
		idsInsertionStmt.Close()
		transaction.Commit()
	}

	settingsInsertionCmd := `INSERT INTO comic_Settings(useDefaultsBits, notifMode, accumCount, delayDuration) VALUES(?, ?, ?, ?);`
	infosInsertionCmd := `INSERT INTO comic_Infos(title, type, status, scanStatus, desc, rating, mature, image) VALUES(?, ?, ?, ?, ?, ?, ?, ?);`
	comicsInsertionCmd := `INSERT INTO comics(infoId, settingsId) VALUES(?, ?);`
	altTitlesInsertionCmd := `INSERT INTO info_AltTitles(title) VALUES(?);`
	altTitlesRelationCmd := `INSERT INTO rel_Info_AltTitles(infoId, titleId) VALUES(?, ?);`
	authorsRelationCmd := `INSERT INTO rel_Info_Authors(infoId, authorId) VALUES(?, ?);`
	artistsRelationCmd := `INSERT INTO rel_Info_Artists(infoId, artistId) VALUES(?, ?);`
	genresRelationCmd := `INSERT INTO rel_Info_Genres(infoId, genreId) VALUES(?, ?);`
	tagsRelationCmd := `INSERT INTO rel_Info_Tags(infoId, tagId) VALUES(?, ?);`
	sourcesInsertionCmd := `INSERT INTO comic_Sources(pluginName, url, markAsRead) VALUES(?, ?, ?);`
	sourcesRelationCmd := `INSERT INTO rel_Comic_Sources(comicId, sourceId) VALUES(?, ?);`
	chaptersInsertionCmd := `INSERT INTO comic_Chapters(identity, alreadyRead) VALUES(?, ?);`
	chaptersRelationCmd := `INSERT INTO rel_Comic_Chapters(comicId, chapterId) VALUES(?, ?);`
	scanlationInsertionCmd := `INSERT INTO chapter_Scanlations(title, lang, pluginName, url) VALUES(?, ?, ?, ?);`
	scanlationRelationCmd := `INSERT INTO rel_Chapter_Scanlations(chapterId, scanlationid) VALUES(?, ?);`
	scanlatorsRelationCmd := `INSERT INTO rel_Scanlations_Scanlators(scanlationId, scanlatorId) VALUES(?, ?);`
	pageLinksInsertionCmd := `INSERT INTO scanlation_PageLinks(url) VALUES(?);`
	pageLinksRelationCmd := `INSERT INTO rel_Scanlations_PageLinks(scanlationId, pageLinkId) VALUES(?, ?);`
	for _, comic := range this {
		transaction, _ := db.Begin()

		lastIdStmt, _ := transaction.Prepare(lastIdCmd)

		var settingsId int
		settingsInsertionStmt, _ := transaction.Prepare(settingsInsertionCmd)
		stts := &comic.Settings
		if !stts.Valid() {
			stts = NewIndividualSettings(LoadGlobalSettings()) //FIXME: use current globals, not saved!
		}
		settingsInsertionStmt.Exec(
			qutils.BoolsToBitfield(stts.UseDefaults), stts.UpdateNotificationMode,
			stts.AccumulativeModeCount, stts.DelayedModeDuration,
		)
		lastIdStmt.QueryRow().Scan(&settingsId)
		settingsInsertionStmt.Close()

		var infoId int
		infoInsertionStmt, _ := transaction.Prepare(infosInsertionCmd)
		inf := &comic.Info
		infoInsertionStmt.Exec(
			inf.Title, inf.Type, inf.Status, inf.ScanlationStatus, inf.Description,
			inf.Rating, inf.Mature, "TODO", //FIXME: get path to thumbnail
		)
		lastIdStmt.QueryRow().Scan(&infoId)
		infoInsertionStmt.Close()

		var comicId int
		comicInsertionStmt, _ := transaction.Prepare(comicsInsertionCmd)
		comicInsertionStmt.Exec(infoId, settingsId)
		lastIdStmt.QueryRow().Scan(&comicId)
		comicInsertionStmt.Close()

		altTitlesInsertionStmt, _ := transaction.Prepare(altTitlesInsertionCmd)
		altTitlesRelationStmt, _ := transaction.Prepare(altTitlesRelationCmd)
		for title, _ := range inf.AltTitles {
			altTitlesInsertionStmt.Exec(title)
			var titleId int
			lastIdStmt.QueryRow().Scan(&titleId)
			altTitlesRelationStmt.Exec(infoId, titleId)
		}
		altTitlesInsertionStmt.Close()
		altTitlesRelationStmt.Close()

		authorsRelationStmt, _ := transaction.Prepare(authorsRelationCmd)
		for _, authorId := range inf.Authors {
			authorId.ExecuteDBStatement(authorsRelationStmt, infoId)
		}
		authorsRelationStmt.Close()

		artistsRelationStmt, _ := transaction.Prepare(artistsRelationCmd)
		for _, artistId := range inf.Artists {
			artistId.ExecuteDBStatement(artistsRelationStmt, infoId)
		}
		artistsRelationStmt.Close()

		genresRelationStmt, _ := transaction.Prepare(genresRelationCmd)
		for genreId := range inf.Genres {
			genreId.ExecuteDBStatement(genresRelationStmt, infoId)
		}
		genresRelationStmt.Close()

		tagsRelationStmt, _ := transaction.Prepare(tagsRelationCmd)
		for tagId := range inf.Categories {
			tagId.ExecuteDBStatement(tagsRelationStmt, infoId)
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
				sc.Language.ExecuteDBStatement(scanlationInsertionStmt, sc.Title, string(sc.PluginName), sc.URL)
				var scanlationId int
				lastIdStmt.QueryRow().Scan(&scanlationId)
				scanlationRelationStmt.Exec(chapterId, scanlationId)

				for _, scanlator := range sc.Scanlators.ToSlice() {
					scanlator.ExecuteDBStatement(scanlatorsRelationStmt, scanlationId)
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
