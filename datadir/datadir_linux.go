package datadir

import (
	"os/user"
	"path/filepath"
)

func init() {
	luser, err := user.Current()
	if err != nil {
		panic("Critical error! Cannot determine user name.")
	}
	dataDir = filepath.Join(luser.HomeDir, ".local", "share", "quasar")
	configDir = filepath.Join(luser.HomeDir, ".config", "quasar")
	logsDir = filepath.Join(dataDir, "logs")
	createDirs()
}
