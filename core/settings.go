package core

import (
	"encoding/json"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
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
	LangName         string
)

type GlobalSettings struct {
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        time.Duration
	MaxConnectionsToHost  uint
	NotificationMode      NotificationMode
	AccumulativeModeCount uint
	DelayedModeDuration   time.Duration
	DownloadsPath         string
	Plugins               map[FetcherPluginName]PluginEnabled
	Languages             map[LangName]LanguageEnabled //TODO: languages validation
	//TODO: default plugin priority?
}

func (this *GlobalSettings) Save() { //TODO: if this == nil, save defaults?
	jsonData, _ := json.MarshalIndent(this.toJSONProxy(), "", "\t")
	WriteConfig(globalConfigFilename, jsonData)
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
		Languages:             this.Languages,
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
		Languages:             map[LangName]LanguageEnabled{LangName(ENGLISH_LANG_NAME): LanguageEnabled(true)},
	}
}

func LoadGlobalSettings() (settings *GlobalSettings, e error) {
	configPath := filepath.Join(datadir.Configs(), globalConfigFilename)
	file, err := os.Open(configPath)
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
		corruptedConfigPath := configPath + ".corrupted"
		os.Remove(corruptedConfigPath)
		os.Rename(configPath, corruptedConfigPath)
		return nil, qerr.NewParse("Error while unmarshaling settings", err, string(jsonData))
	}
	if proxy.MaxConnectionsToHost > 10 {
		proxy.MaxConnectionsToHost = 10 //bigger values seem to trigger a DDoS protection, so clamp for now
	}
	settings = proxy.toSettings()

	return
}

type globalSettingsJSONProxy struct {
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        SplitDuration
	MaxConnectionsToHost  uint
	ValidModeValues       []string //can't have comments in JSON, make it a dummy value instead
	NotificationMode      string
	AccumulativeModeCount uint
	DelayedModeDuration   SplitDuration
	DownloadsPath         string
	Plugins               map[FetcherPluginName]PluginEnabled `json:"PluginsEnabled"`
	Languages             map[LangName]LanguageEnabled        `json:"LangsEnabled"`
}

func (this *globalSettingsJSONProxy) toSettings() *GlobalSettings {
	return &GlobalSettings{
		FetchOnStartup:        this.FetchOnStartup,
		IntervalFetching:      this.IntervalFetching,
		FetchFrequency:        this.FetchFrequency.ToDuration(),
		MaxConnectionsToHost:  this.MaxConnectionsToHost,
		NotificationMode:      NotificationModeFromString(this.NotificationMode),
		AccumulativeModeCount: this.AccumulativeModeCount,
		DelayedModeDuration:   this.DelayedModeDuration.ToDuration(),
		DownloadsPath:         this.DownloadsPath,
		Plugins:               this.Plugins,
		Languages:             this.Languages,
	}
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
	AccumulativeModeCount uint
	DelayedModeDuration   time.Duration
	DownloadPath          string
}

func (this *IndividualSettings) Valid() bool {
	return len(this.OverrideDefaults) != 0
}

func NewIndividualSettings(defaults *GlobalSettings) *IndividualSettings {
	return &IndividualSettings{
		OverrideDefaults:      make([]bool, bitlength_comic),
		FetchOnStartup:        defaults.FetchOnStartup,
		IntervalFetching:      defaults.IntervalFetching,
		NotificationMode:      defaults.NotificationMode,
		AccumulativeModeCount: defaults.AccumulativeModeCount,
		DelayedModeDuration:   defaults.DelayedModeDuration,
		DownloadPath:          defaults.DownloadsPath,
	}
}

var downloadsPath string

func init() {
	luser, _ := user.Current()                                          //how can this even fail o_O
	downloadsPath = filepath.Join(luser.HomeDir, "Downloads", "Quasar") //TODO: get default path (e.g. use xdg on Linux)
}

const globalConfigFilename = "config.json"

func WriteConfig(filename string, data []byte) {
	configDir := datadir.Configs()
	os.MkdirAll(configDir, os.ModeDir|0755)
	ioutil.WriteFile(filepath.Join(configDir, filename), data, 0644)
}

func ReadConfig(filename string) (contents []byte, err error) {
	file, err := os.Open(filepath.Join(datadir.Configs(), filename))
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
	MaxConnectionsToHost  uint
	NotificationMode      NotificationMode
	AccumulativeModeCount uint
	DelayedModeDuration   time.Duration
	Languages             map[LangName]LanguageEnabled
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
