package datadir

import "os"

var (
	dataDir   string
	configDir string
	logsDir   string
)

func createDirs() {
	os.MkdirAll(dataDir, os.ModeDir|0755)
	os.MkdirAll(configDir, os.ModeDir|0755)
	os.MkdirAll(logsDir, os.ModeDir|0755)
}

func Path() string {
	return dataDir
}

func Configs() string {
	return configDir
}

func Logs() string {
	return logsDir
}

func OverrideLogs(newPath string) {
	logsDir = newPath
}
