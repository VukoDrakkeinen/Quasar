package core

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"time"

	"database/sql/driver"
	"errors"
	"fmt"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"github.com/VukoDrakkeinen/qml"
)

const (
	hoursPerDay Duration = 24
	daysPerWeek          = 7
	hourTime             = Duration(time.Hour)
	dayTime              = hourTime * hoursPerDay
	weekTime             = dayTime * daysPerWeek

	//weeksPerMonth               = 4 //uniform
	//monthsPerYear               = 13 //4*7*13 = 364; 13th month is... Nonuary?
	//monthTime = weekTime * weeksPerMonth
	//yearTime = monthTime * monthsPerYear
)
const (
	Immediate NotificationMode = iota
	Accumulative
	Delayed
)

var bitlength_common = reflect.TypeOf(CommonSettings{}).NumField() - 1
var bitlength_comiccfg = reflect.TypeOf(ComicConfig{}).NumField() - 1 + bitlength_common
var bitlength_sourcecfg = reflect.TypeOf(SourceConfig{}).NumField() - 1 + bitlength_common

func init() {
	if bitlength_comiccfg > 64 || bitlength_sourcecfg > 64 {
		panic("Too many booleans to encode in a single 64-bit field!")
	}
}

type (
	NotificationMode int
	PluginEnabled    bool
	LanguageEnabled  bool
	LangName         string
)

type CommonSettings struct {
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        Duration
	NotificationMode      NotificationMode
	AccumulativeModeCount uint
	DelayedModeDuration   Duration
}

type GlobalSettings struct {
	CommonSettings
	IAmADirtyLeecher     bool
	MaxConnectionsToHost uint
	DownloadsPath        string
	Plugins              map[SourceId]PluginEnabled   `json:"PluginsEnabled"`
	Languages            map[LangName]LanguageEnabled `json:"LangsEnabled"` //TODO: languages validation
	Ignore_JSONComment0  []string                     `json:"#ValidNModeValues"`
	//TODO: default plugin priority?
	//TODO: color scheme
	//TODO: user-agent
}

func (this *GlobalSettings) Save() { //TODO: if this == nil, save defaults?
	this.Ignore_JSONComment0 = NotificationModeValueNames()
	jsonData, _ := json.MarshalIndent(this, "", "\t")
	WriteConfig(globalConfigFilename, jsonData)
}

func NewGlobalSettings() *GlobalSettings {
	return &GlobalSettings{
		IAmADirtyLeecher: false,
		CommonSettings: CommonSettings{
			FetchOnStartup:        true,
			IntervalFetching:      true,
			FetchFrequency:        Duration(time.Hour * 3),
			NotificationMode:      Immediate,
			AccumulativeModeCount: 10,
			DelayedModeDuration:   Duration(time.Hour * 24 * 7),
		},
		MaxConnectionsToHost: 10,
		DownloadsPath:        downloadsPath,
		Plugins:              make(map[SourceId]PluginEnabled),
		Languages:            map[LangName]LanguageEnabled{LangName(ENGLISH_LANG_NAME): LanguageEnabled(true)},
	}
}

func LoadGlobalSettings() (settings *GlobalSettings, e error) {
	jsonData, err := ReadConfig(globalConfigFilename)
	configPath := filepath.Join(datadir.Configs(), globalConfigFilename)
	if os.IsNotExist(err) {
		settings = NewGlobalSettings()
		settings.Save()
		return settings, nil
	} else if err != nil {
		return nil, err
	}

	err = json.Unmarshal(jsonData, &settings)
	if err != nil {
		corruptedConfigPath := configPath + ".corrupted"
		os.Remove(corruptedConfigPath)
		os.Rename(configPath, corruptedConfigPath)
		return nil, qerr.NewParse("Error while unmarshaling settings", err, string(jsonData))
	}

	if settings.MaxConnectionsToHost < 1 { //TODO: better validation
		settings.MaxConnectionsToHost = 1
		qlog.Log(qlog.Warning, "Invalid number of maximum connections! Can't be zero.")
	} else if settings.MaxConnectionsToHost > 5 {
		settings.MaxConnectionsToHost = 5 //bigger values seem to trigger a DDoS protection, so clamp for now
		qlog.Log(qlog.Warning, "More than 5 simultaneous connections may trigger a DDoS protection!")
	}

	return settings, nil
}

type splitDuration struct {
	Hours uint8 `json:"hours"`
	Days  uint8 `json:"days"`
	Weeks uint8 `json:"weeks"`
	//Months uint8 `json:"months"`
	//Years  uint8 `json:"years"`
}

func (this splitDuration) duration() (d Duration) {
	d += Duration(this.Hours) * hourTime
	d += Duration(this.Days) * dayTime
	d += Duration(this.Weeks) * weekTime
	return
}

type Duration time.Duration

func (this Duration) Split() (s splitDuration) {
	s.Weeks, this = uint8(this/weekTime), this%weekTime
	s.Days, this = uint8(this/dayTime), this%dayTime
	s.Hours = uint8(this / hourTime)
	return s
}

func (this Duration) MarshalJSON() ([]byte, error) {
	data, err := json.Marshal(this.Split())
	return data, err
}

func (this *Duration) UnmarshalJSON(data []byte) error {
	var split splitDuration
	err := json.Unmarshal(data, &split)
	if err != nil {
		return err
	}
	*this = split.duration()
	return nil
}

func (this Duration) MarshalQML() interface{} {
	return this.Split()
}

func (this *Duration) UnmarshalQML(data interface{}) (err error) {
	defer func() { //todo: get rid of this defer, use normal errors
		r := recover()
		if r != nil {
			err = r.(error)
		}
	}()
	switch data := data.(type) {
	case splitDuration:
		*this = data.duration()
	case *qml.Map:
		var split splitDuration
		data.Unmarshal(&split)
		*this = split.duration()
	default:
		_ = data.(Duration)
	}
	return err
}

func (this Duration) Value() (driver.Value, error) {
	return int64(this), nil
}

func (this *Duration) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be an int64, got %T!)", *this, src))
	}
	*this = Duration(n)
	return nil
}

type ComicConfig struct {
	OverrideDefaults []bool
	CommonSettings
	DownloadPath string
}

func (this *ComicConfig) Valid() bool {
	return len(this.OverrideDefaults) != 0
}

func NewComicConfig(defaults *GlobalSettings) ComicConfig {
	return ComicConfig{
		OverrideDefaults: make([]bool, bitlength_comiccfg),
		CommonSettings: CommonSettings{
			FetchOnStartup:        defaults.FetchOnStartup,
			FetchFrequency:        defaults.FetchFrequency,
			IntervalFetching:      defaults.IntervalFetching,
			NotificationMode:      defaults.NotificationMode,
			AccumulativeModeCount: defaults.AccumulativeModeCount,
			DelayedModeDuration:   defaults.DelayedModeDuration,
		},
		DownloadPath: defaults.DownloadsPath,
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

type SourceConfig struct {
	OverrideDefaults []bool
	CommonSettings
	MaxConnectionsToHost uint
	Languages            map[LangName]LanguageEnabled
}

func NewSourceConfig(defaults *GlobalSettings) SourceConfig {
	return SourceConfig{
		OverrideDefaults: make([]bool, bitlength_sourcecfg),
		CommonSettings: CommonSettings{
			FetchOnStartup:        defaults.FetchOnStartup,
			IntervalFetching:      defaults.IntervalFetching,
			FetchFrequency:        defaults.FetchFrequency,
			NotificationMode:      defaults.NotificationMode,
			AccumulativeModeCount: defaults.AccumulativeModeCount,
			DelayedModeDuration:   defaults.DelayedModeDuration,
		},
		MaxConnectionsToHost: defaults.MaxConnectionsToHost,
		Languages:            defaults.Languages,
	}
}
