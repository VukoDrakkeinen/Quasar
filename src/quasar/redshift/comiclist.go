package redshift

import (
	"database/sql"
	"quasar/qutils"
	"quasar/redshift/qdb"
)

type ComicList []Comic

func (this *ComicList) saveToDB() {
	db := qdb.DB() //TODO: error out on nil
	createSettings := `
	CREATE TABLE IF NOT EXISTS comic.Settings (
	  id INTEGER PRIMARY KEY,
	  useDefaultsBits INTEGER NOT NULL,
	  notifMode INTEGER,
	  accumCount INTEGER,
	  delayDuration INTEGER,
	  downloadsPath TEXT
	);
	`
	createInfos := `
	CREATE TABLE IF NOT EXISTS comic.Infos (
	id INTEGER PRIMARY KEY,
	title TEXT NOT NULL,
	altTitlesTable TEXT NOT NULL,
	authorsTable
	`
	_, err := db.Exec(createSettings)
	if err != nil {
		//TODO: log error
		return
	}

	transaction, err := db.Begin()
	if err != nil {
		//TODO: log error
	}
	insertSettings := `
	INSERT OR REPLACE INTO comicSettings(
	  useDefaultsBits,
	  notifMode,
	  acumCount,
	  delayDuration
	  )
	values(?, ?, ?, ?)
	`
	settingsInsertion, err := transaction.Prepare(insertSettings)
	if err != nil {
		//TODO: log error
	}
	defer settingsInsertion.Close()
	for _, comic := range this {
		settings := &comic.Settings
		_, err = settingsInsertion.Exec(qutils.BoolsToBitfield(settings.UseDefaults), settings.UpdateNotificationMode,
			settings.AccumulativeModeCount, settings.DelayedModeDuration)
		if err != nil {
			//TODO: log error
		}
	}
	transaction.Commit()
}
