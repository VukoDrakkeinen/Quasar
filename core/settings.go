package core

import (
	"encoding/json"
	. "github.com/Quasar/core/idsdict"
	"github.com/Quasar/qutils/qerr"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"time"
)

const (
	hoursPerDay time.Duration = 24
	daysPerWeek               = 7
	//weeksPerMonth               = 4 //uniform
	//monthsPerYear               = 13 //4*7*13 = 364; 13th month is Nonuary :P
	dayTime  = time.Hour * hoursPerDay
	weekTime = dayTime * daysPerWeek
	//monthTime = weekTime * weeksPerMonth
	//yearTime = monthTime * monthsPerYear
)
const (
	Immediate NotificationMode = iota
	Accumulative
	Delayed
)

var bitlength_comic = reflect.TypeOf(IndividualSettings{}).NumField() - 1
var bitlength_plugin = reflect.TypeOf(PerPluginSettings{}).NumField() - 1

type BitfieldType int

const (
	ComicSettings BitfieldType = iota
	PluginSettings
)

func Bitlength(typ BitfieldType) int {
	switch typ {
	case ComicSettings:
		return bitlength_comic
	case PluginSettings:
		return bitlength_plugin
	default:
		return -1
	}
}

type (
	NotificationMode int
	PluginEnabled    bool
	LanguageEnabled  bool
)

type GlobalSettings struct {
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        time.Duration
	MaxConnectionsToHost  int
	NotificationMode      NotificationMode
	AccumulativeModeCount int
	DelayedModeDuration   time.Duration
	DownloadsPath         string
	Plugins               map[FetcherPluginName]PluginEnabled
	Languages             map[LangId]LanguageEnabled
	//TODO: default plugin priority?
}

func (this *GlobalSettings) Save() { //TODO: if this == nil, save defaults?
	jsonData, _ := json.MarshalIndent(this.toJSONProxy(), "", "\t")
	WriteConfig(globalConfigFile, jsonData)
}

func (this *GlobalSettings) toJSONProxy() *globalSettingsJSONProxy {
	proxy := &globalSettingsJSONProxy{
		FetchOnStartup:        this.FetchOnStartup,
		IntervalFetching:      this.IntervalFetching,
		FetchFrequency:        DurationToSplit(this.FetchFrequency),
		MaxConnectionsToHost:  this.MaxConnectionsToHost,
		ValidModeValues:       NotificationModeValueNames(),
		NotificationMode:      this.NotificationMode.String(),
		AccumulativeModeCount: this.AccumulativeModeCount,
		DelayedModeDuration:   DurationToSplit(this.DelayedModeDuration),
		DownloadsPath:         this.DownloadsPath,
		Plugins:               this.Plugins,
		Languages:             make(map[string]LanguageEnabled),
	}
	for id, status := range this.Languages {
		proxy.Languages[Langs.NameOf(id)] = status
	}
	return proxy
}

func NewGlobalSettings() *GlobalSettings {
	return &GlobalSettings{
		FetchOnStartup:        true,
		IntervalFetching:      true,
		FetchFrequency:        time.Duration(time.Hour * 3),
		MaxConnectionsToHost:  10,
		NotificationMode:      Immediate,
		AccumulativeModeCount: 10,
		DelayedModeDuration:   time.Duration(time.Hour * 24 * 7),
		DownloadsPath:         downloadsPath,
		Plugins:               make(map[FetcherPluginName]PluginEnabled),
		Languages:             map[LangId]LanguageEnabled{ENGLISH_LANG(): LanguageEnabled(true)},
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
	var proxy globalSettingsJSONProxy = *NewGlobalSettings().toJSONProxy()
	err = json.Unmarshal(jsonData, &proxy)
	if err != nil {
		return nil, qerr.NewParse("Error while unmarshaling settings", err, string(jsonData))
	}
	settings = proxy.toSettings()

	return
}

type globalSettingsJSONProxy struct {
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        SplitDuration
	MaxConnectionsToHost  int
	ValidModeValues       []string //can't have comments in JSON, make it a dummy value instead
	NotificationMode      string
	AccumulativeModeCount int
	DelayedModeDuration   SplitDuration
	DownloadsPath         string
	Plugins               map[FetcherPluginName]PluginEnabled `json:"PluginsEnabled"`
	Languages             map[string]LanguageEnabled          `json:"LangsEnabled"`
}

func (this *globalSettingsJSONProxy) toSettings() *GlobalSettings {
	settings := &GlobalSettings{
		FetchOnStartup:        this.FetchOnStartup,
		IntervalFetching:      this.IntervalFetching,
		FetchFrequency:        this.FetchFrequency.ToDuration(),
		MaxConnectionsToHost:  this.MaxConnectionsToHost,
		NotificationMode:      NotificationModeFromString(this.NotificationMode),
		AccumulativeModeCount: this.AccumulativeModeCount,
		DelayedModeDuration:   this.DelayedModeDuration.ToDuration(),
		DownloadsPath:         this.DownloadsPath,
		Plugins:               this.Plugins,
		Languages:             make(map[LangId]LanguageEnabled, len(this.Languages)),
	}
	for lang, status := range this.Languages {
		settings.Languages[Langs.Id(lang)] = status
	}
	return settings
}

type SplitDuration struct {
	Hours uint8 `json:"hours"`
	Days  uint8 `json:"days"`
	Weeks uint8 `json:"weeks"`
	//Months uint8 `json:"months"`
	//Years  uint8 `json:"years"`
}

func (this SplitDuration) ToDuration() (d time.Duration) {
	d += time.Duration(this.Hours) * time.Hour
	d += time.Duration(this.Days) * dayTime
	d += time.Duration(this.Weeks) * weekTime
	return
}

func DurationToSplit(d time.Duration) (s SplitDuration) {
	s.Weeks, d = uint8(d/weekTime), d%weekTime
	s.Days, d = uint8(d/dayTime), d%dayTime
	s.Hours = uint8(d / time.Hour)
	return
}

type IndividualSettings struct { //TODO: rename -> PerComicSettings
	OverrideDefaults      []bool
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        time.Duration
	NotificationMode      NotificationMode
	AccumulativeModeCount int
	DelayedModeDuration   time.Duration
	DownloadPath          string
}

func (this *IndividualSettings) Valid() bool {
	return len(this.OverrideDefaults) != 0
}

func NewIndividualSettings(defaults *GlobalSettings) *IndividualSettings {
	return &IndividualSettings{
		OverrideDefaults:      make([]bool, bitlength_comic),
		NotificationMode:      defaults.NotificationMode,
		AccumulativeModeCount: defaults.AccumulativeModeCount,
		DelayedModeDuration:   defaults.DelayedModeDuration,
		DownloadPath:          defaults.DownloadsPath,
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

type PerPluginSettings struct {
	OverrideDefaults      []bool
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        time.Duration
	MaxConnectionsToHost  int
	NotificationMode      NotificationMode
	AccumulativeModeCount int
	DelayedModeDuration   time.Duration
	Languages             map[LangId]LanguageEnabled
}

func NewPerPluginSettings(defaults *GlobalSettings) PerPluginSettings {
	return PerPluginSettings{
		OverrideDefaults:      make([]bool, bitlength_plugin),
		FetchOnStartup:        defaults.FetchOnStartup,
		IntervalFetching:      defaults.IntervalFetching,
		FetchFrequency:        defaults.FetchFrequency,
		MaxConnectionsToHost:  defaults.MaxConnectionsToHost,
		NotificationMode:      defaults.NotificationMode,
		AccumulativeModeCount: defaults.AccumulativeModeCount,
		DelayedModeDuration:   defaults.DelayedModeDuration,
		Languages:             defaults.Languages,
	}
}
