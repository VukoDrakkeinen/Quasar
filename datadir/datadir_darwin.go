package datadir

import (
	"os/user"
	"path/filepath"
)

func init() {
	luser, err := user.Current()
	if err != nil { //TODO: Console errors aren't too user-friendly
		panic("Critical error! Cannot determine user name.")
	}
	dataDir = filepath.Join(luser.HomeDir, "Library", "Application Support", "Quasar")
	configDir = filepath.Join(luser.HomeDir, "Library", "Preferences", "Quasar")
	logsDir = filepath.Join(luser.HomeDir, "Library", "Logs", "Quasar")
	//plugins: ~/Library/Application Support/
	createDirs()
}
