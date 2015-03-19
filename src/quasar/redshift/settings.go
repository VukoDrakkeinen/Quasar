package redshift

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"quasar/qutils/qerr"
	. "quasar/redshift/idsdict"
	"reflect"
	"time"
)

const (
	hoursPerDay  time.Duration = 24
	hoursPerWeek               = 7 * hoursPerDay
	weekTime                   = time.Hour * hoursPerWeek
	dayTime                    = time.Hour * hoursPerDay
)

type UpdateNotificationMode int

const (
	OnLaunch UpdateNotificationMode = iota
	Accumulative
	Delayed
	Manual
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

func (this *GlobalSettings) Save() {
	jsonData, _ := json.MarshalIndent(this.toJSONProxy(), "", "\t")
	WriteConfig(globalConfigFile, jsonData)
}

func (this *GlobalSettings) toJSONProxy() *globalSettingsJSONProxy {
	proxy := &globalSettingsJSONProxy{
		ValidModeValues:               UpdateNotificationModeValueNames(),
		DefaultUpdateNotificationMode: this.DefaultUpdateNotificationMode.String(),
		DefaultAccumulativeModeCount:  this.DefaultAccumulativeModeCount,
		DefaultDelayedModeDuration:    durationToSplit(this.DefaultDelayedModeDuration),
		DefaultDownloadsPath:          this.DefaultDownloadsPath,
		Plugins:                       this.Plugins,
		Languages:                     make(map[string]LanguageEnabled),
	}
	for id, status := range this.Languages {
		proxy.Languages[Langs.NameOf(id)] = status
	}
	return proxy
}

func NewGlobalSettings() *GlobalSettings {
	return &GlobalSettings{
		DefaultUpdateNotificationMode: OnLaunch,
		DefaultAccumulativeModeCount:  10,
		DefaultDelayedModeDuration:    time.Duration(time.Hour * 24 * 7),
		DefaultDownloadsPath:          downloadsPath,
		Plugins:                       make(map[FetcherPluginName]PluginEnabled),
		Languages:                     map[LangId]LanguageEnabled{ENGLISH_LANG(): LanguageEnabled(true)},
	}
}

func LoadGlobalSettings() (settings *GlobalSettings, e error) {
	file, err := os.Open(filepath.Join(configDir, globalConfigFile))
	defer file.Close()
	if os.IsNotExist(err) {
		settings = NewGlobalSettings()
		settings.Save()
		return
	} else if err != nil {
		return nil, err
	}
	jsonData, _ := ioutil.ReadAll(file)
	var proxy globalSettingsJSONProxy
	err = json.Unmarshal(jsonData, &proxy)
	if err != nil {
		return nil, qerr.NewParse("Error while unmarshaling settings", err, string(jsonData))
	}
	settings = proxy.toSettings()

	return
}

type globalSettingsJSONProxy struct {
	ValidModeValues               []string                            //can't have comments in JSON, make it a dummy value instead
	DefaultUpdateNotificationMode string                              `json:"UpdateNotificationMode"`
	DefaultAccumulativeModeCount  int                                 `json:"AccumulativeModeCount"`
	DefaultDelayedModeDuration    splitDuration                       `json:"DelayedModeDuration"`
	DefaultDownloadsPath          string                              `json:"DownloadsPath"`
	Plugins                       map[FetcherPluginName]PluginEnabled `json:"PluginsEnabled"`
	Languages                     map[string]LanguageEnabled          `json:"LangsEnabled"`
}

func (this *globalSettingsJSONProxy) toSettings() *GlobalSettings {
	settings := &GlobalSettings{
		DefaultUpdateNotificationMode: UpdateNotificationModeFromString(this.DefaultUpdateNotificationMode),
		DefaultAccumulativeModeCount:  this.DefaultAccumulativeModeCount,
		DefaultDelayedModeDuration:    this.DefaultDelayedModeDuration.toDuration(),
		DefaultDownloadsPath:          this.DefaultDownloadsPath,
		Plugins:                       this.Plugins,
		Languages:                     make(map[LangId]LanguageEnabled, len(this.Languages)),
	}
	for lang, status := range this.Languages {
		settings.Languages[Langs.Id(lang)] = status
	}
	return settings
}

type splitDuration struct {
	Hours time.Duration `json:"hours"`
	Days  time.Duration `json:"days"`
	Weeks time.Duration `json:"weeks"`
}

func (this *splitDuration) toDuration() (d time.Duration) {
	d += this.Hours * time.Hour
	d += this.Days * dayTime
	d += this.Weeks * weekTime
	return
}

func durationToSplit(d time.Duration) (s splitDuration) {
	s.Weeks, d = d/weekTime, d%weekTime
	s.Days, d = d/dayTime, d%dayTime
	s.Hours = d / time.Hour
	return
}

type IndividualSettings struct {
	UseDefaults            []bool
	UpdateNotificationMode UpdateNotificationMode
	AccumulativeModeCount  int
	DelayedModeDuration    time.Duration
	DownloadPath           string
}

func (this *IndividualSettings) Valid() bool {
	return len(this.UseDefaults) != 0
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

//TODO:
//XP: + \Local Settings\Application Data\Quasar\
//Win: + \AppData\Local\Quasar\
//OSX: + /Library/Application Support/Quasar/
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
