package gui

import (
	"github.com/Quasar/core"
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

func (this *coreConnector) AddComic(settingsDatas *qml.List, sources *qml.Map) {
	var data []int
	var plugins []string
	var urls []string
	var marks []bool
	var sourcesMap map[string]*qml.List
	settingsDatas.Convert(&data)
	sources.Convert(&sourcesMap)
	sourcesMap["plugins"].Convert(&plugins)
	sourcesMap["urls"].Convert(&urls)
	sourcesMap["marks"].Convert(&marks)
	settings := core.NewIndividualSettings(this.list.Fetcher().Settings())
	settings.NotificationMode = core.NotificationMode(data[0])
	settings.AccumulativeModeCount = data[1]
	settings.DelayedModeDuration = core.SplitDuration{uint8(data[2]), uint8(data[3]), uint8(data[4])}.ToDuration()
	comic := core.NewComic(*settings)
	for i := 0; i < len(plugins); i++ {
		source := core.UpdateSource{
			PluginName: core.FetcherPluginName(plugins[i]),
			URL:        urls[i],
			MarkAsRead: marks[i],
		}
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
