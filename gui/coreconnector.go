package gui

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"gopkg.in/qml.v1"
	"reflect"
)

func NewCoreConnector(list *core.ComicList) *coreConnector {
	return &coreConnector{
		list: list,
	}
}

type coreConnector struct {
	list *core.ComicList
}

func (this *coreConnector) PluginNames() (names *[]core.FetcherPluginName, humanReadableNames *[]string) {
	pluginNames, hrNames := this.list.Fetcher().Plugins() //TODO?: sorted data
	return &pluginNames, &hrNames
}

func (this *coreConnector) PluginAutodetect(url string) (pluginName core.FetcherPluginName) {
	fetcherPluginName, _ := this.list.Fetcher().PluginNameFromURL(url)
	return fetcherPluginName
}

func (this *coreConnector) AddComic(settingsObj, dmDuration *qml.Map, sourcesList *qml.List) {
	settings := core.NewIndividualSettings(this.list.Fetcher().Settings())

	var splitDuration core.SplitDuration
	dmDuration.Unmarshal(&splitDuration)
	settingsObj.Unmarshal(settings)
	settings.DelayedModeDuration = splitDuration.ToDuration()

	var sources []*qml.Map
	sourcesList.Convert(&sources)

	comic := core.NewComic(*settings)
	for _, sourceObj := range sources {
		var source core.UpdateSource
		sourceObj.Unmarshal(&source)
		comic.AddSource(source)
	}

	go func() {
		this.list.Fetcher().DownloadComicInfoFor(comic)
		this.list.AddComics([]*core.Comic{comic})
		this.list.ScheduleComicFetches()
	}()
}

type temporaryNeuteredGlobalSettings struct { //TODO: remove? how? go-to-qml/qml-to-go type converters?
	FetchOnStartup        bool
	IntervalFetching      bool
	FetchFrequency        core.SplitDuration
	MaxConnectionsToHost  uint
	NotificationMode      core.NotificationMode
	AccumulativeModeCount uint
	DelayedModeDuration   core.SplitDuration
	DownloadsPath         string
	Plugins               map[core.FetcherPluginName]core.PluginEnabled //FIXME: causes a crash (unhashable value)
	Languages             map[idsdict.LangId]core.LanguageEnabled       //this too
}

func (this *coreConnector) GlobalSettings() *temporaryNeuteredGlobalSettings {
	settings := this.list.Fetcher().Settings()

	return &temporaryNeuteredGlobalSettings{
		FetchOnStartup:        settings.FetchOnStartup,
		IntervalFetching:      settings.IntervalFetching,
		FetchFrequency:        core.DurationToSplit(settings.FetchFrequency),
		MaxConnectionsToHost:  settings.MaxConnectionsToHost,
		NotificationMode:      settings.NotificationMode,
		AccumulativeModeCount: settings.AccumulativeModeCount,
		DelayedModeDuration:   core.DurationToSplit(settings.DelayedModeDuration),
		DownloadsPath:         settings.DownloadsPath,
		Plugins:               settings.Plugins,
		Languages:             settings.Languages,
	}
}

func (this *coreConnector) SetGlobalSettings(settingsObj, dmDuration *qml.Map, fetchFrequency *qml.Map) {
	settings := this.list.Fetcher().Settings()

	var splitDuration, splitFrequency core.SplitDuration
	dmDuration.Unmarshal(&splitDuration)
	fetchFrequency.Unmarshal(&splitFrequency)
	settingsObj.Unmarshal(settings)
	settings.DelayedModeDuration = splitDuration.ToDuration()
	settings.FetchFrequency = splitFrequency.ToDuration()
}

func (this *coreConnector) DefaultGlobalSettings() *temporaryNeuteredGlobalSettings {
	settings := this.list.Fetcher().Settings() //we still need it for some data
	defaults := core.NewGlobalSettings()

	neutered := &temporaryNeuteredGlobalSettings{
		FetchOnStartup:        defaults.FetchOnStartup,
		IntervalFetching:      defaults.IntervalFetching,
		FetchFrequency:        core.DurationToSplit(defaults.FetchFrequency),
		MaxConnectionsToHost:  defaults.MaxConnectionsToHost,
		NotificationMode:      defaults.NotificationMode,
		AccumulativeModeCount: defaults.AccumulativeModeCount,
		DelayedModeDuration:   core.DurationToSplit(defaults.DelayedModeDuration),
		DownloadsPath:         defaults.DownloadsPath,
		Plugins:               settings.Plugins,
		Languages:             settings.Languages,
	}
	for p := range neutered.Plugins {
		neutered.Plugins[p] = core.PluginEnabled(true)
	}
	return neutered
}

func (this *coreConnector) ComicSettings(idx int) *temporaryNeuteredGlobalSettings {
	settings := this.list.GetComic(idx).Settings()

	return &temporaryNeuteredGlobalSettings{
		NotificationMode:      settings.NotificationMode,
		AccumulativeModeCount: settings.AccumulativeModeCount,
		DelayedModeDuration:   core.DurationToSplit(settings.DelayedModeDuration),
	}
}

func (this *coreConnector) SetComicSettingsAndSources(comicIdx int, settingsObj, dmDuration *qml.Map, sourcesList *qml.List) {
	comic := this.list.GetComic(comicIdx)
	settings := comic.Settings()

	prevSources := comic.Sources()

	var splitDuration core.SplitDuration
	dmDuration.Unmarshal(&splitDuration)
	settingsObj.Unmarshal(&settings)
	settings.DelayedModeDuration = splitDuration.ToDuration()
	comic.SetSettings(settings)

	var sources []*qml.Map
	sourcesList.Convert(&sources) //TODO: update comic after data changes
	for i, sourceObj := range sources {
		var source core.UpdateSource
		sourceObj.Unmarshal(&source)
		comic.AddSourceAt(i, source)
	}

	if !reflect.DeepEqual(comic.Sources(), prevSources) {
		go this.list.UpdateComic(comicIdx)
	}

}

func (this *coreConnector) ComicSources(comicIdx int) *[]core.UpdateSource {
	csources := this.list.GetComic(comicIdx).Sources()
	return &csources
}

func (this *coreConnector) UpdateComics(comicIndices *qml.List) {
	var ids []int
	comicIndices.Convert(&ids)
	for _, i := range ids {
		go this.list.UpdateComic(i)
	}
}
