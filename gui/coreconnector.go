package gui

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"gopkg.in/qml.v1"
	"unsafe"
)

func NewCoreConnector(list *core.ComicList) *coreConnector {
	return &coreConnector{
		list: list,
	}
}

type coreConnector struct {
	list *core.ComicList
}

func (this *coreConnector) PluginNames() (names *[]string, humanReadableNames *[]string) {
	pluginNames, hrNames := this.list.Fetcher().Plugins() //TODO: important! sorted data!
	//screw FetcherPluginName -> string conversion, we'll have to convert it into C++ data anyway
	return (*[]string)(unsafe.Pointer(&pluginNames)), &hrNames
}

func (this *coreConnector) PluginAutodetect(url string) (pluginName string) {
	fetcherPluginName, _ := this.list.Fetcher().PluginNameFromURL(url)
	return *(*string)(unsafe.Pointer(&fetcherPluginName))
}

func (this *coreConnector) AddComic(settingsObj *qml.Map, sourcesList *qml.List) {
	var neuteredSettings temporaryNeuteredGlobalSettings
	settingsObj.Unmarshal(&neuteredSettings)
	settings := core.NewIndividualSettings(this.list.Fetcher().Settings())
	settings.NotificationMode = core.NotificationMode(neuteredSettings.NotificationMode)
	settings.AccumulativeModeCount = neuteredSettings.AccumulativeModeCount
	settings.DelayedModeDuration = neuteredSettings.DelayedModeDuration.ToDuration()

	var sources []*qml.Map
	sourcesList.Convert(&sources)

	comic := core.NewComic(*settings)
	for _, sourceObj := range sources {
		var source core.UpdateSource
		sourceObj.Unmarshal(&source)
		comic.AddSource(source)
	}

	go func() { //TODO: show progress
		this.list.Fetcher().DownloadComicInfoFor(comic)
		this.list.AddComics([]*core.Comic{comic})
		this.list.ScheduleComicFetches() //TODO: just one
	}()
}

type temporaryNeuteredGlobalSettings struct {
	NotificationMode      int
	AccumulativeModeCount int
	DelayedModeDuration   core.SplitDuration
}

func (this *coreConnector) GlobalSettings() *temporaryNeuteredGlobalSettings {
	settings := this.list.Fetcher().Settings()
	return &temporaryNeuteredGlobalSettings{
		NotificationMode:      int(settings.NotificationMode),
		AccumulativeModeCount: settings.AccumulativeModeCount,
		DelayedModeDuration:   core.DurationToSplit(settings.DelayedModeDuration),
	}
}

func (this *coreConnector) ComicSettings(idx int) *temporaryNeuteredGlobalSettings {
	settings := this.list.GetComic(idx).Settings()
	return &temporaryNeuteredGlobalSettings{
		NotificationMode:      int(settings.NotificationMode),
		AccumulativeModeCount: settings.AccumulativeModeCount,
		DelayedModeDuration:   core.DurationToSplit(settings.DelayedModeDuration),
	}
}

func (this *coreConnector) SetComicSettingsAndSources(comicIdx int, settingsObj *qml.Map, sourcesList *qml.List) {
	comic := this.list.GetComic(comicIdx)

	var neuteredSettings temporaryNeuteredGlobalSettings
	settingsObj.Unmarshal(&neuteredSettings)
	settings := comic.Settings()
	settings.NotificationMode = core.NotificationMode(neuteredSettings.NotificationMode)
	settings.AccumulativeModeCount = neuteredSettings.AccumulativeModeCount
	settings.DelayedModeDuration = neuteredSettings.DelayedModeDuration.ToDuration()
	comic.SetSettings(settings)

	var sources []*qml.Map
	sourcesList.Convert(&sources)
	for i, sourceObj := range sources {
		var source core.UpdateSource
		sourceObj.Unmarshal(&source)
		comic.AddSourceAt(i, source)
	}
}

type updateSource struct {
	PluginName string
	URL        string
	MarkAsRead bool
}

func (this *coreConnector) ComicSources(comicIdx int) *[]updateSource {
	csources := this.list.GetComic(comicIdx).Sources()
	sources := make([]updateSource, 0, len(csources))
	for _, source := range csources {
		sources = append(sources, updateSource{string(source.PluginName), source.URL, source.MarkAsRead})
	}
	return &sources
}
