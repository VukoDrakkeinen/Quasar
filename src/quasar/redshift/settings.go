package redshift

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	. "quasar/redshift/idsdict"
	"reflect"
	"time"
)

type PluginEnabled bool
type LanguageEnabled bool
type GlobalSettings struct {
	DefaultUpdateNotificationMode UpdateNotificationMode
	DefaultAccumulativeModeCount  int
	DefaultDelayedModeDuration    time.Duration
	DefaultDownloadsPath          string
	Plugins                       map[FetcherPluginName]PluginEnabled
	Languages                     map[LangId]LanguageEnabled
	//TODO: default plugin priority?
}

type globalSettingsJSONProxy struct {
	DefaultUpdateNotificationMode UpdateNotificationMode              `json:"updateNotificationMode"`
	DefaultAccumulativeModeCount  int                                 `json:"accumulativeModeCount"`
	DefaultDelayedModeDuration    time.Duration                       `json:"delayedModeDuration"` //TODO: serialize as hours:days:months:years?
	DefaultDownloadsPath          string                              `json:"downloadsPath"`
	Plugins                       map[FetcherPluginName]PluginEnabled `json:"pluginsEnabled"`
	Languages                     map[string]LanguageEnabled          `json:"langsEnabled"`
}

func (this *GlobalSettings) toJSONProxy() *globalSettingsJSONProxy {
	proxy := &globalSettingsJSONProxy{
		DefaultUpdateNotificationMode: this.DefaultUpdateNotificationMode,
		DefaultAccumulativeModeCount:  this.DefaultAccumulativeModeCount,
		DefaultDelayedModeDuration:    this.DefaultDelayedModeDuration,
		DefaultDownloadsPath:          this.DefaultDownloadsPath,
		Plugins:                       this.Plugins,
	}
	proxy.Languages = make(map[string]LanguageEnabled)
	for id, status := range this.Languages {
		proxy.Languages[Langs.NameOf(id)] = status
	}
	return proxy
}

func (this *globalSettingsJSONProxy) toSettings() *GlobalSettings {
	settings := &GlobalSettings{
		DefaultUpdateNotificationMode: this.DefaultUpdateNotificationMode,
		DefaultAccumulativeModeCount:  this.DefaultAccumulativeModeCount,
		DefaultDelayedModeDuration:    this.DefaultDelayedModeDuration,
		DefaultDownloadsPath:          this.DefaultDownloadsPath,
		Plugins:                       this.Plugins,
	}
	for lang, status := range this.Languages {
		settings.Languages[Langs.Id(lang)] = status
	}
	return settings
}

func NewGlobalSettings() *GlobalSettings {
	return &GlobalSettings{
		DefaultUpdateNotificationMode: Immediate,
		DefaultAccumulativeModeCount:  10,
		DefaultDelayedModeDuration:    time.Duration(time.Hour * 24 * 7),
		DefaultDownloadsPath:          downloadsPath,
		Plugins:                       make(map[FetcherPluginName]PluginEnabled),
		Languages:                     map[LangId]LanguageEnabled{ENGLISH_LANG(): LanguageEnabled(true)},
	}
}

type IndividualSettings struct {
	UseDefaults            []bool
	UpdateNotificationMode UpdateNotificationMode
	AccumulativeModeCount  int
	DelayedModeDuration    time.Duration
	DownloadPath           string
}

func initDefaults() []bool {
	ret := make([]bool, 0, reflect.TypeOf(IndividualSettings{}).NumField()-1)
	for i := 0; i < cap(ret); i++ {
		ret = append(ret, true)
	}
	return ret
}

func NewIndividualSettings(defaults *GlobalSettings) *IndividualSettings {
	return &IndividualSettings{
		UseDefaults:            initDefaults(),
		UpdateNotificationMode: defaults.DefaultUpdateNotificationMode,
		AccumulativeModeCount:  defaults.DefaultAccumulativeModeCount,
		DelayedModeDuration:    defaults.DefaultDelayedModeDuration,
		DownloadPath:           defaults.DefaultDownloadsPath,
	}
}

func (this *IndividualSettings) Valid() bool {
	return len(this.UseDefaults) != 0
}

//TODO:
//XP: + \Local Settings\Application Data\Quasar\
//Win: + \AppData\Local\Quasar\
//OSX: + /Library/Application Support/Quasar/
//Linux: + /.config/quasar/
var configDir string
var downloadsPath string

func init() {
	luser, _ := user.Current() //how can this even fail o_O
	configDir = filepath.Join(luser.HomeDir, ".config", "quasar")
	downloadsPath = filepath.Join(luser.HomeDir, "Downloads", "Quasar")
}

const globalConfigFile = "config.json"

func WriteConfig(filename string, data []byte) {
	os.MkdirAll(configDir, os.ModeDir|0755)
	ioutil.WriteFile(filepath.Join(configDir, filename), data, 0644)
}

func ReadConfig(filename string) (contents []byte, err error) {
	file, err := os.Open(filepath.Join(configDir, filename))
	defer file.Close()
	if err != nil {
		return
	}
	contents, err = ioutil.ReadAll(file)
	return
}

func (this *GlobalSettings) Save() {
	jsonData, _ := json.MarshalIndent(this.toJSONProxy(), "", "\t")
	WriteConfig(globalConfigFile, jsonData)
}

func LoadGlobalSettings() (settings *GlobalSettings) { //TODO: refactor
	file, err := os.Open(filepath.Join(configDir, globalConfigFile))
	defer file.Close()
	if os.IsNotExist(err) {
		settings = NewGlobalSettings()
		settings.Save()
	} else if err != nil {
		//TODO: handle errors
		panic("Cannot load global settings: " + err.Error())
	} else {
		jsonData, _ := ioutil.ReadAll(file) //TODO: handle errors
		var proxy *globalSettingsJSONProxy
		err := json.Unmarshal(jsonData, proxy)
		if err != nil {
			//TODO: log error
			settings = NewGlobalSettings()
		} else {
			settings = proxy.toSettings()
		}
	}
	return
}

type UpdateNotificationMode int

const (
	Immediate UpdateNotificationMode = iota
	Accumulative
	Delayed
	Manual
)
