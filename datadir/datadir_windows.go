package datadir

import (
	"os"
	"os/user"
	"path/filepath"
	"syscall"
)

func init() { //Get your shit together Windows, look how short other OS's implementations are...
	appData := os.Getenv("APPDATA")
	localAppData := os.Getenv("LOCALAPPDATA")

	if localAppData != "" {
		dataDir = filepath.Join(localAppData, "Quasar") //Vista and higher
	} else {
		dataDir = filepath.Join(appData, "Quasar") //XP
	}
	configDir = filepath.Join(appData, "Quasar")

	if appData == "" {
		luser, err := user.Current()
		if err != nil { //TODO: console error on Windows...? Should probably show a popup message.
			panic("Critical error! Cannot determine application data folders locations and user name.")
		}
		dll := syscall.MustLoadDLL("kernel32.dll")
		proc := dll.MustFindProc("GetVersion")
		ver, _, _ := proc.Call()                 //version is 32 bits consisting of build(16):minor(8):major(8)
		if majorVer := byte(ver); majorVer < 6 { //Vista is 6.0, XP is 5.1
			dataDir = filepath.Join(luser.HomeDir, "Local Settings", "Application Data", "Quasar") //XP
			configDir = dataDir
		} else {
			dataDir = filepath.Join(luser.HomeDir, "AppData", "Local", "Quasar") //Vista and higher
			configDir = filepath.Join(luser.HomeDir, "AppData", "Roaming", "Quasar")
		}
	}

	logsDir = filepath.Join(dataDir, "logs")
	createDirs()
}
