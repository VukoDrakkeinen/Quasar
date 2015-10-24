package core

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"errors"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/math"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type correctiveSlice struct {
	identities     []ChapterIdentity
	chapters       []Chapter
	missingVolumes bool
}

func (this correctiveSlice) Len() int {
	return math.Min(len(this.identities), len(this.chapters))
}

func (this correctiveSlice) Less(i, j int) bool {
	return this.identities[i].Less(this.identities[j])
}

func (this correctiveSlice) Swap(i, j int) {
	this.identities[i], this.identities[j] = this.identities[j], this.identities[i]
	this.chapters[i], this.chapters[j] = this.chapters[j], this.chapters[i]
}

func (this correctiveSlice) Correct() {
	if this.missingVolumes {
		maxLen := this.Len()
		for i := 0; i < maxLen; i++ {
			if this.identities[i].Version != 1 && i+1 < maxLen {
				for j := i + 1; j < maxLen; j++ {
					if this.identities[j].n()&^0xFFFF != this.identities[i].n()&^0xFFFF { //compare only vol:num
						//sort.Sort(correctiveSlice{this.identities[i:j], this.chapters[i:j]})
						qutils.Reverse(correctiveSlice{this.identities[i:j], this.chapters[i:j], this.missingVolumes})
						i = j
						break
					}
				}
			}
		}
		prevVol := byte(1)
		for i := range this.identities {
			if this.identities[i].Volume == 0 {
				this.identities[i].Volume = prevVol
			}
			prevVol = this.identities[i].Volume
		}
	}
	sort.Sort(this)
}

type fetcher struct { //TODO: handle missing plugin errors gracefully
	plugins     map[FetcherPluginName]FetcherPlugin
	webClient   *http.Client
	settings    *GlobalSettings
	cache       *DataCache
	connsToHost map[string]uint
	maxConns    map[FetcherPluginName]uint
	cond        *sync.Cond
}

func NewFetcher(settings *GlobalSettings, plugins ...FetcherPlugin) *fetcher {
	fet := &fetcher{
		plugins: make(map[FetcherPluginName]FetcherPlugin),
		webClient: &http.Client{
			CheckRedirect: nil,
		},
		settings:    settings,
		cache:       NewDataCache(),
		connsToHost: make(map[string]uint, 10),
		maxConns:    make(map[FetcherPluginName]uint, 10),
		cond:        sync.NewCond(&sync.Mutex{}),
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
		plugin.SetSettings(NewPerPluginSettings(this.settings)) //TODO
		for _, lang := range plugin.Languages() {
			langName := LangName(lang)
			this.settings.Languages[langName] = this.settings.Languages[langName] || LanguageEnabled(false)
		}
		this.settings.Plugins[name] = PluginEnabled(true)
		successes = append(successes, true) //TODO?
		replaced = append(replaced, pluginReplaced)
	}
	return
}

func (this *fetcher) PluginLimitsUpdated(pluginName FetcherPluginName, maxConns uint) {
	this.cond.L.Lock()
	if maxConns != 0 {
		this.maxConns[pluginName] = maxConns
	} else {
		this.maxConns[pluginName] = this.settings.MaxConnectionsToHost
	}
	this.cond.L.Unlock()
}

func (this *fetcher) PluginProvidedLanguages() (langNames []string) {
	langSet := make(map[string]struct{})
	for _, plugin := range this.plugins {
		for _, lang := range plugin.Languages() {
			if _, duplicate := langSet[lang]; !duplicate {
				langSet[lang] = struct{}{}
				langNames = append(langNames, lang)
			}
		}
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
		wg.Add(1)
		go func(pluginName FetcherPluginName) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					this.pluginPanicked(pluginName, err)
				}
			}()
			comic.SetInfo(*comic.Info().MergeWith(this.plugins[pluginName].fetchComicInfo(comic)))
		}(source.PluginName)
	}
	wg.Wait()
}

func (this *fetcher) getConnectionPermit(pluginName FetcherPluginName, host string) {
	this.cond.L.Lock()
	for this.connsToHost[host] == this.maxConns[pluginName] {
		this.cond.Wait()
	}
	this.connsToHost[host]++
	this.cond.L.Unlock()
	return
}

func (this *fetcher) giveupConnectionPermit(host string) {
	this.cond.L.Lock()
	this.connsToHost[host]--
	this.cond.L.Unlock()
	this.cond.Signal()
}

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

			if clen := response.ContentLength; clen > 0 {
				data = make([]byte, response.ContentLength)
				_, errChan := qutils.BackgroundCopy(body, bytes.NewBuffer(data)) //TODO: propagate progress info further
				err = <-errChan
			} else {
				data, err = ioutil.ReadAll(body)
			}
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
		case 502, 503, 504: //TODO: be smarter here (try avoiding triggering the DDoS protection by spreading retries)
			waitTime := 10*time.Second + (time.Duration(rand.Intn(100)+1)*200+time.Duration(rand.Intn(2))*100)*time.Millisecond
			qlog.Logf(qlog.Warning, "Connection to %s rejected (%d/5). Retrying in %v", parsedUrl.Host, i, waitTime)
			time.Sleep(waitTime)
			continue
		default:
			return nil, errors.New(`Unhandled response status code "` + response.Status + `" received!`)
		}
	}
	return nil, errors.New(`Maximum amount of retries exceeded!`)
}

func (this *fetcher) pluginPanicked(offender FetcherPluginName, err interface{}) {
	qlog.Log(qlog.Error, "Plugin", string(offender), "panicked!", "Message:", err)
	qlog.Logf(qlog.Error, "\n%s\n", qutils.Stack())
	this.settings.Plugins[offender] = PluginEnabled(false)
}

func (this *fetcher) DownloadChapterListFor(comic *Comic) { //TODO: skipAllowed boolean (optimisation, download only last page to update existing list, the suggestion may be disregarded) - only some plugins
	var wg sync.WaitGroup
	for _, source := range comic.Sources() {
		wg.Add(1)
		go func(pluginName FetcherPluginName) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					this.pluginPanicked(pluginName, err)
				}
			}()

			if plugin, success := this.plugins[pluginName]; success && plugin.Capabilities().ProvidesMetadata {
				identities, chapters, missingVolumes := plugin.fetchChapterList(comic)
				//some plugins return ChapterIdentities with no Volume data, correct it, then sort
				correctiveSlice{identities, chapters, missingVolumes}.Correct()
				comic.AddMultipleChapters(identities, chapters, false)
			}
		}(source.PluginName)
	}
	wg.Wait()
}

func (this *fetcher) DownloadPageLinksFor(comic *Comic, chapterIndex, scanlationIndex int) (success bool) {
	var offender FetcherPluginName //TODO: notify view
	defer func() {
		if err := recover(); err != nil {
			this.pluginPanicked(offender, err)
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

func getChapterDirPath(settings *GlobalSettings, comic *Comic, ci ChapterIdentity) string { //FIXME: brittle
	path := filepath.Join(settings.DownloadsPath, comic.Info().MainTitle, ci.Stringify())
	os.MkdirAll(path, os.ModeDir|0755)
	return path
}

func (this *fetcher) DownloadPages(comic *Comic, chapterIndex, scanlationIndex int) { //FIXME: temporary implementation
	println("dp")
	var offender FetcherPluginName //TODO: notify view
	defer func() {
		if err := recover(); err != nil {
			this.pluginPanicked(offender, err)
		}
	}()

	chapter, ci := comic.GetChapter(chapterIndex)
	scanlation := chapter.Scanlation(scanlationIndex)
	if plugin, success := this.plugins[scanlation.PluginName]; success && plugin.Capabilities().ProvidesData {
		offender = scanlation.PluginName
		if len(scanlation.PageLinks) == 0 {
			this.DownloadPageLinksFor(comic, chapterIndex, scanlationIndex)
		}
		dirpath := getChapterDirPath(this.settings, comic, ci)
		for _, link := range scanlation.PageLinks {
			filename := link[strings.LastIndex(link, "/")+1:]
			file, err := os.Create(dirpath + "/" + filename)
			if err != nil {
				println(err)
				continue
			}
			data, err := this.DownloadData(offender, link, false) //TODO: try to avoid copying
			if err != nil {
				println(err)
				continue
			}
			file.Write(data)
			file.Close()
		}
		println("done")
	}
	return
}

func (this *fetcher) PluginNameFromURL(url string) (FetcherPluginName, error) {
	var offender FetcherPluginName
	defer func() {
		if err := recover(); err != nil {
			this.pluginPanicked(offender, err)
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

func (this *fetcher) FindComic(title string) []comicSearchResult {
	var wg sync.WaitGroup
	sharedResults := make(chan []comicSearchResult, 1)
	sharedResults <- make([]comicSearchResult, 0, 2)
	for name, plugin := range this.plugins {
		wg.Add(1)
		go func(pluginName FetcherPluginName, plugin FetcherPlugin) {
			defer wg.Done()
			defer func() {
				if err := recover(); err != nil {
					this.pluginPanicked(pluginName, err)
				}
			}()

			url := plugin.findComicURL(title) //TODO
			if url == "" {
				return
			}

			results := <-sharedResults
			defer func() { sharedResults <- results }()
			results = append(results, comicSearchResult{PluginName: pluginName, Title: title, URL: url}) //TODO
		}(name, plugin)
	}
	wg.Wait()
	return <-sharedResults
}

func (this *fetcher) FindComicAdvanced(title string) {} //TODO: more params

type comicSearchResult struct {
	PluginName FetcherPluginName
	Title      string
	URL        string
	//optional:
	Authors    string
	Rating     uint16
	Views      int
	Year       int
	LastUpdate time.Time
}
