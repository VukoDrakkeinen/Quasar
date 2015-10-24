package gui

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/eventq"
	"gopkg.in/qml.v1"
	"reflect"
	"sort"
)

var (
	ChaptersMarked = eventq.NewEventType()
)

func NewCoreConnector(list *core.ComicList) *coreConnector {
	return &coreConnector{list}
}

type coreConnector struct {
	list *core.ComicList
}

func (this *coreConnector) PluginNames() (names []core.FetcherPluginName, humanReadableNames []string) {
	return this.list.Fetcher().Plugins() //TODO: map?
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
		this.list.AddComics(comic)
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
	Plugins               map[core.FetcherPluginName]core.PluginEnabled
	Languages             map[core.LangName]core.LanguageEnabled
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

func (this *coreConnector) SetGlobalSettings(settingsObj, dmDuration, fetchFrequency *qml.Map) {
	settings := this.list.Fetcher().Settings()

	var splitDuration, splitFrequency core.SplitDuration //TODO: type converters
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

func (this *coreConnector) ComicSources(comicIdx int) []core.UpdateSource {
	return this.list.GetComic(comicIdx).Sources()
}

func (this *coreConnector) UpdateComics(comicIndices *qml.List) {
	var ids []int
	comicIndices.Convert(&ids)
	for _, i := range ids {
		go this.list.UpdateComic(i)
	}
}

func (this *coreConnector) MarkAsRead(comicIdx int, chapterIndicesList *qml.List, read bool) {
	comic := this.list.GetComic(comicIdx)
	var chapterIndices []int
	chapterIndicesList.Convert(&chapterIndices)
	sort.Ints(chapterIndices)
	chapters := make([]core.Chapter, 0, len(chapterIndices))
	identities := make(core.ChapterIdentitiesSlice, 0, len(chapterIndices)) //TODO: will be quite slow
	last := -2
	selections := make([][2]int, 0, len(chapterIndices))
	for _, i := range chapterIndices { //consider modifying in-place (pointers!)
		if i == last {
			continue
		} else if i != last+1 {
			selections = append(selections, [2]int{i, 1}) //add row, count 1
		} else {
			selections[len(selections)-1][1]++ //increment count
		}
		chapter, id := comic.GetChapter(i)
		chapter.AlreadyRead = read
		chapters = append(chapters, chapter)
		identities = append(identities, id)
		last = i
	}

	comic.AddMultipleChapters(identities, chapters, true)
	eventq.Event(ChaptersMarked, comicIdx, selections)
}

func (this *coreConnector) DownloadPages(comicIdx int, chapterIndicesList, scanlationIndicesList *qml.List) {
	comic := this.list.GetComic(comicIdx)
	var chapterIndices, scanlationIndices []int
	chapterIndicesList.Convert(&chapterIndices)
	scanlationIndicesList.Convert(&scanlationIndices)
	for i := range chapterIndices {
		go func() {
			println("step1")
			this.list.Fetcher().DownloadPageLinksFor(comic, chapterIndices[i], scanlationIndices[i])
			println("step2")
			this.list.Fetcher().DownloadPages(comic, chapterIndices[i], scanlationIndices[i])
			println("profit")
		}()
		//TODO: don't download needlessly
	}
	//TODO: show progress
}

func (this *coreConnector) GetQueuedChapter(comicIdx int) (chapterIdx int) {
	comic := this.list.GetComic(comicIdx)
	return comic.QueuedChapter()
}

func (this *coreConnector) GetLastReadChapter(comicIdx int) (chapterIdx int) {
	comic := this.list.GetComic(comicIdx)
	return comic.LastReadChapter()
}
