package datadir

import (
	"os"
	"os/user"
	"path/filepath"
)

var dataDir string

func init() {
	luser, _ := user.Current()
	dataDir = filepath.Join(luser.HomeDir, ".local", "share", "quasar")
	os.MkdirAll(dataDir, os.ModeDir|0755)
}

func Path() string {
	return dataDir
}
