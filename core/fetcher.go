package core

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"io"
	"io/ioutil"
	"net/http"
	neturl "net/url"
	"sort"
	"sync"
	"time"
)

type correctiveSlice struct {
	identities []ChapterIdentity
	chapters   []Chapter
}

func (this correctiveSlice) Len() int {
	if ilen := len(this.identities); ilen == len(this.chapters) {
		return ilen
	} else {
		return 0
	}
}

func (this correctiveSlice) Less(i, j int) bool {
	ident1 := this.identities[i]
	ident2 := this.identities[j]
	if ident1.Volume != 0 && ident2.Volume != 0 {
		return ident1.Less(ident2)
	} else {
		return ident1.MajorNum < ident2.MajorNum ||
			(ident1.MajorNum == ident2.MajorNum && ident1.MinorNum < ident2.MinorNum)
	}
}

func (this correctiveSlice) Swap(i, j int) {
	this.identities[i], this.identities[j] = this.identities[j], this.identities[i]
	this.chapters[i], this.chapters[j] = this.chapters[j], this.chapters[i]
}

type fetcher struct { //TODO: handle missing plugin errors gracefully
	plugins     map[FetcherPluginName]FetcherPlugin
	webClient   *http.Client
	settings    *GlobalSettings
	cache       *DataCache
	connLimiter map[string]chan struct{}
	connLimits  map[FetcherPluginName]int
	notifyView  func(work func())
}

func NewFetcher(settings *GlobalSettings, notifyViewFunc func(work func()), plugins ...FetcherPlugin) *fetcher {
	fet := &fetcher{
		plugins: make(map[FetcherPluginName]FetcherPlugin),
		webClient: &http.Client{
			CheckRedirect: nil,
		},
		settings:    settings,
		cache:       NewDataCache(),
		connLimiter: make(map[string]chan struct{}),
		connLimits:  make(map[FetcherPluginName]int),
		notifyView:  notifyViewFunc,
	}
	if fet.settings == nil {
		fet.settings = NewGlobalSettings()
	}
	fet.RegisterPlugins(plugins...)
	return fet
}

func (this *fetcher) RegisterPlugins(plugins ...FetcherPlugin) (successes, replaced []bool) {
	for _, plugin := range plugins {
		name := plugin.PluginName()
		oldPlugin, pluginReplaced := this.plugins[name]
		if pluginReplaced {
			oldPlugin.setFetcher(nil)
		}
		this.plugins[name] = plugin
		plugin.setFetcher(this)
		Langs.AssignIds(plugin.Languages())
		plugin.SetSettings(NewPerPluginSettings(this.settings)) //TODO
		this.connLimits[name] = plugin.Settings().MaxConnectionsToHost
		this.settings.Plugins[name] = PluginEnabled(true)
		successes = append(successes, true) //TODO?
		replaced = append(replaced, pluginReplaced)
	}
	return
}

func (this *fetcher) Plugins() (names []FetcherPluginName, humanReadableNames []string) {
	for pluginName, plugin := range this.plugins {
		names = append(names, pluginName)
		humanReadableNames = append(humanReadableNames, plugin.HumanReadableName())
	}
	return
}

func (this *fetcher) DownloadComicInfoFor(comic *Comic) {
	var wg sync.WaitGroup
	for _, source := range comic.Sources() {
		offender := source.PluginName
		wg.Add(1)
		go func() {
			defer func() {
				offender := offender
				if err := recover(); err != nil {
					qlog.Log(qlog.Error, "Plugin", string(offender), "panicked!", err)
					this.settings.Plugins[offender] = PluginEnabled(false)
				}
			}()
			comic.SetInfo(*comic.Info().MergeWith(this.plugins[source.PluginName].fetchComicInfo(comic)))
			wg.Done()
		}()
	}
	wg.Wait()
}

func (this *fetcher) getConnectionPermit(pluginName FetcherPluginName, host string) {
	if c, ok := this.connLimiter[host]; ok {
		if maxConns := this.connLimits[pluginName]; cap(c) != maxConns {
			oldLen := len(c)
			this.connLimiter[host] = make(chan struct{}, maxConns)
			c = this.connLimiter[host]
			for i := 0; i < oldLen; i++ {
				c <- struct{}{}
			}
		}
		<-c
	} else {
		maxConns := this.connLimits[pluginName]
		if maxConns == 0 {
			maxConns = this.settings.MaxConnectionsToHost
			if maxConns == 0 {
				maxConns = 1
				this.settings.MaxConnectionsToHost = 1
				this.settings.Save()
				qlog.Log(qlog.Warning, "Invalid number of maximum connections! Can't be zero.")
			}
		}

		this.connLimiter[host] = make(chan struct{}, maxConns)
		c = this.connLimiter[host]
		for i := 0; i < maxConns; i++ {
			c <- struct{}{}
		}
		<-c
	}
	return
}

func (this *fetcher) giveupConnectionPermit(host string) {
	this.connLimiter[host] <- struct{}{}
}

//TODO: parallelization?
func (this *fetcher) DownloadData(pluginName FetcherPluginName, url string, forceCaching bool) (data []byte, err error) {
	if data, ok := this.cache.Get(url); ok {
		return data, err
	}

	parsedUrl, err := neturl.Parse(url)
	if err != nil {
		return []byte{}, qerr.Chain("Unable to find host in url", err)
	}

	this.getConnectionPermit(pluginName, parsedUrl.Host)
	defer this.giveupConnectionPermit(parsedUrl.Host)

	for i := 0; i < 5; i++ { //TODO: configurable amount of retries?
		request, err := http.NewRequest("GET", url, nil)
		request.Header.Set("User-Agent", `Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:38.0) Gecko/20100101 Firefox/38.0`) //TODO: configurable
		request.Header.Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
		request.Header.Set("Accept-Encoding", `gzip, deflate`)
		request.Header.Set("Accept-Language", `en-GB,en;q=0.5`)
		request.Header.Set("DNT", `1`)
		request.Header.Set("Connection", `keep-alive`)
		request.Header.Set("Cache-Control", `max-age=0`)
		response, err := this.webClient.Do(request)
		if err != nil {
			return nil, err
		}

		switch response.StatusCode {
		case 200:
			var body io.ReadCloser = response.Body
			defer body.Close()
			switch response.Header.Get("Content-Encoding") {
			case "gzip":
				body, err = gzip.NewReader(body)
				if err != nil {
					return []byte{}, qerr.Chain("Corrupted GZIP data!", err)
				}
				defer body.Close()
			case "deflate":
				body = flate.NewReader(body)
				defer body.Close()
			}

			data, err = ioutil.ReadAll(body)
			if err != nil {
				return []byte{}, err
			} else if forceCaching {
				this.cache.Add(url, data, time.Duration(0))
			}
			return data, err
		case 301, 302:
			url = response.Header.Get("Location")
			parsedRedirect, err := neturl.Parse(url)
			if err != nil {
				return []byte{}, qerr.Chain("Unable to find host in url", err)
			}
			if parsedUrl.Host != parsedRedirect.Host {
				this.getConnectionPermit(pluginName, parsedRedirect.Host)
				defer this.giveupConnectionPermit(parsedRedirect.Host)
			}
			continue
		case 502, 503, 504:
			time.Sleep(2 * time.Second)
			continue
		default:
			return nil, errors.New(`Unhandled response status code "` + response.Status + `" received!`)
		}
	}
	return nil, errors.New(`Maximum amount of retries exceeded!`)
}

func (this *fetcher) DownloadChapterListFor(comic *Comic) { //TODO: skipAllowed boolean (optimisation, download only last page to update existing list, the suggestion may be disregarded) - only some plugins
	this.notifyView(func() {
		var wg sync.WaitGroup
		for _, source := range comic.Sources() {
			offender := source.PluginName
			wg.Add(1)
			go func() {
				defer func() {
					if err := recover(); err != nil {
						offender := offender
						qlog.Log(qlog.Error, "Plugin", string(offender), "panicked!", err)
						this.settings.Plugins[offender] = PluginEnabled(false)
					}
				}()

				identities, chapters, missingVolumes := this.plugins[source.PluginName].fetchChapterList(comic)
				if missingVolumes { //some plugins return ChapterIdentities with no Volume data, correct it
					correctiveSlice := correctiveSlice{identities, chapters}
					sort.Sort(correctiveSlice)
					prevVol := byte(1)
					for i := range correctiveSlice.identities {
						if correctiveSlice.identities[i].Volume == 0 {
							correctiveSlice.identities[i].Volume = prevVol
						}
						prevVol = correctiveSlice.identities[i].Volume
					}
				}
				comic.AddMultipleChapters(identities, chapters)
				wg.Done()
			}()
		}
		wg.Wait()
	})
}

func (this *fetcher) DownloadPageLinksFor(comic *Comic, chapterIndex, scanlationIndex int) (success bool) {
	var offender FetcherPluginName
	defer func() {
		if err := recover(); err != nil {
			qlog.Log(qlog.Error, "Plugin", string(offender), "panicked!", err)
			this.settings.Plugins[offender] = PluginEnabled(false)
		}
	}()

	chapter, identity := comic.GetChapter(chapterIndex)
	scanlation := chapter.Scanlation(scanlationIndex)
	if plugin, success := this.plugins[scanlation.PluginName]; success && plugin.Capabilities().ProvidesData {
		offender = scanlation.PluginName
		links := plugin.fetchChapterPageLinks(scanlation.URL)
		scanlation.PageLinks = links
		chapter.AddScanlation(scanlation)    //reinsert after modifying
		comic.AddChapter(identity, &chapter) //reinsert //TODO: use pointers instead?
	}
	return
}

func (this *fetcher) PluginNameFromURL(url string) (FetcherPluginName, error) {
	var offender FetcherPluginName
	defer func() {
		if err := recover(); err != nil {
			qlog.Log(qlog.Error, "Plugin", string(offender), "panicked!", err)
			this.settings.Plugins[offender] = PluginEnabled(false)
		}
	}()

	for pluginName, plugin := range this.plugins {
		offender = pluginName
		if plugin.IsURLValid(url) {
			return pluginName, nil
		}
	}
	return "", errors.New("Plugin autodetect failed!")
}

func (this *fetcher) Settings() *GlobalSettings {
	return this.settings
}

func (this *fetcher) TestFind(comic *Comic, pluginName FetcherPluginName, comicTitle string) {
	plugin := this.plugins[pluginName]
	urlFound := plugin.findComicURL(comicTitle)
	if urlFound != "" {
		comic.AddSource(UpdateSource{
			PluginName: pluginName,
			URL:        urlFound,
			MarkAsRead: false,
		})
	}
}
