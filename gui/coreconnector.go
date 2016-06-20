package gui

import (
	"reflect"
	"sort"

	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/eventq"
)

var (
	ChaptersMarked = eventq.NewEventType()
)

func NewCoreConnector(list *core.ComicList) *coreConnector {
	return &coreConnector{
		list:      list,
		Messenger: eventq.NewMessenger(),
	}
}

type coreConnector struct {
	list *core.ComicList

	eventq.Messenger
}

func (this *coreConnector) PluginNames() (names []core.SourceId, humanReadableNames []string) {
	return this.list.Fetcher().Plugins() //TODO: map?
}

func (this *coreConnector) PluginAutodetect(url string) (pluginName core.SourceId) {
	pluginName, _ = this.list.Fetcher().PluginNameFromURL(url)
	return pluginName
}

func (this *coreConnector) AddComic(config core.ComicConfig, sources []core.SourceLink) {
	tempCfgMore := core.NewComicConfig(this.list.Fetcher().Settings()) //todo
	tempCfgMore.NotificationMode = config.NotificationMode
	tempCfgMore.AccumulativeModeCount = config.AccumulativeModeCount
	tempCfgMore.DelayedModeDuration = config.DelayedModeDuration

	comic := core.NewComic(tempCfgMore)
	for _, source := range sources {
		comic.AddSourceLink(source)
	}

	go func() {
		this.list.Fetcher().FetchComicInfoFor(&comic)
		this.list.AddComics(comic)
		this.list.ScheduleComicFetches()
	}()
}

func (this *coreConnector) GlobalSettings() *core.GlobalSettings {
	return this.list.Fetcher().Settings()
}

func (this *coreConnector) SetGlobalSettings(settings core.GlobalSettings) {
	*this.list.Fetcher().Settings() = settings
}

func (this *coreConnector) DefaultGlobalSettings() *core.GlobalSettings {
	settings := this.list.Fetcher().Settings() //we still need it for some data
	defaults := core.NewGlobalSettings()
	defaults.Plugins = make(map[core.SourceId]core.PluginEnabled, len(settings.Plugins))
	defaults.Languages = settings.Languages
	for p := range defaults.Plugins {
		defaults.Plugins[p] = core.PluginEnabled(true)
	}
	return defaults
}

func (this *coreConnector) ComicConfig(idx int) *core.ComicConfig { //TODO: why the pointer? doesn't need to be assignable
	cfg := this.list.GetComic(idx).Config()
	return &cfg
}

func (this *coreConnector) SetComicConfigAndSources(comicIdx int, config core.ComicConfig, sources []core.SourceLink) {
	comic := this.list.GetComic(comicIdx)

	tempCfgMore := comic.Config() //todo
	tempCfgMore.NotificationMode = config.NotificationMode
	tempCfgMore.AccumulativeModeCount = config.AccumulativeModeCount
	tempCfgMore.DelayedModeDuration = config.DelayedModeDuration
	comic.SetConfig(tempCfgMore)

	prevSources := comic.SourceLinks()

	for i, source := range sources {
		comic.AddSourceLinkAt(i, source)
	}

	if !reflect.DeepEqual(comic.SourceLinks(), prevSources) {
		go this.list.UpdateComic(comicIdx)
	}

}

func (this *coreConnector) ComicSources(comicIdx int) []core.SourceLink { //todo: finish the massive renaming
	return this.list.GetComic(comicIdx).SourceLinks()
}

func (this *coreConnector) UpdateComics(comicIndices []int) {
	for _, i := range comicIndices {
		go this.list.UpdateComic(i)
	}
}

func (this *coreConnector) MarkAsRead(comicIdx int, chapterIndices []int, read bool) {
	comic := this.list.GetComic(comicIdx)
	sort.Ints(chapterIndices)
	last := -2
	selections := make([][2]int, 0, len(chapterIndices))
	for _, i := range chapterIndices {
		if i == last {
			continue
		} else if i != last+1 {
			selections = append(selections, [2]int{i, 1}) //add row, count 1
		} else {
			selections[len(selections)-1][1]++ //increment count
		}
		chapter, _ := comic.Chapter(i)
		chapter.MarkedRead = read
		last = i
	}

	this.Event(ChaptersMarked, comicIdx, selections)
}

func (this *coreConnector) DownloadPages(comicIdx int, chapterIndices, scanlationIndices []int) {
	comic := this.list.GetComic(comicIdx)
	for i := range chapterIndices {
		go func() {
			i := i
			println("step1")
			this.list.Fetcher().FetchPageLinksFor(comic, chapterIndices[i], scanlationIndices[i])
			println("step2")
			this.list.Fetcher().DownloadPages(comic, chapterIndices[i], scanlationIndices[i])
			println("profit")
		}()
		//TODO: don't download needlessly
	}
	//TODO: show progress
}

func (this *coreConnector) GetQueuedChapter(comicIdx int) (chapterIdx int) {
	return this.list.GetComic(comicIdx).QueuedChapter()
}

func (this *coreConnector) GetLastReadChapter(comicIdx int) (chapterIdx int) {
	return this.list.GetComic(comicIdx).LastReadChapter()
}
