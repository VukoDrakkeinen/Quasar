package redshift

import (
	"compress/flate"
	"compress/gzip"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	. "quasar/redshift/idsdict"
	"sort"
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

//TODO: scheduler

type fetcher struct { //TODO: handle missing plugin errors gracefully
	plugins   map[FetcherPluginName]FetcherPlugin
	webClient *http.Client
	settings  *GlobalSettings
	cache     *DataCache
}

func NewFetcher(settings *GlobalSettings, plugins ...FetcherPlugin) *fetcher {
	fet := &fetcher{
		plugins: make(map[FetcherPluginName]FetcherPlugin),
		webClient: &http.Client{
			CheckRedirect: nil, //TODO: write the redirect handling function
		},
		settings: settings,
		cache:    NewDataCache(),
	}
	if fet.settings == nil {
		fet.settings = NewGlobalSettings()
	}
	for _, plugin := range plugins {
		fet.RegisterPlugin(plugin)
	}
	return fet
}

func (this *fetcher) RegisterPlugin(plugin FetcherPlugin) (success, replaced bool) {
	name := plugin.PluginName()
	oldPlugin, replaced := this.plugins[name]
	if replaced {
		oldPlugin.setFetcher(nil)
	}
	this.plugins[name] = plugin
	plugin.setFetcher(this)
	Langs.AssignIds(plugin.Languages())
	success = true //TODO?
	return
}

func (this *fetcher) DownloadComicInfoFor(comic *Comic) {
	for _, source := range comic.Sources() {
		comic.Info.MergeWith(this.plugins[source.PluginName].fetchComicInfo(comic))
	}
}

//TODO: parallelization?
func (this *fetcher) DownloadData(url string, forceCaching bool) []byte {
	if data, ok := this.cache.Get(url); ok {
		return data
	}
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Set("User-Agent", `Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:38.0) Gecko/20100101 Firefox/38.0`) //TODO: configurable
	request.Header.Set("Accept", `text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8`)
	request.Header.Set("Accept-Encoding", `gzip, deflate`)
	request.Header.Set("Accept-Language", `en-GB,en;q=0.5`)
	request.Header.Set("DNT", `1`)
	request.Header.Set("Connection", `keep-alive`)
	request.Header.Set("Cache-Control", `max-age=0`)
	response, _ := this.webClient.Do(request)

	switch response.StatusCode {
	case 200: //OK, continue
	case 301, 302:
		return this.DownloadData(response.Header.Get("Location"), forceCaching)
	case 502, 503, 504:
		time.Sleep(2 * time.Second)
		return this.DownloadData(url, forceCaching)
	default:
		return []byte{} //,errors.New(`Unhandled response status code "` + response.Status + `" received!`) TODO: return error?
	}

	var body io.ReadCloser = response.Body
	defer body.Close()
	switch response.Header.Get("Content-Encoding") {
	case "gzip":
		body, err = gzip.NewReader(body)
		if err != nil {
			return []byte{} //,qerr.Chain("Corrupted GZIP data!", err)
		}
		defer body.Close()
	case "deflate":
		body = flate.NewReader(body)
		defer body.Close()
	}

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return []byte{}
		panic("Error in ioutil.ReadAll()")
	} else if forceCaching {
		this.cache.Add(url, data, time.Duration(0))
	}
	return data //,nil
}

func (this *fetcher) DownloadChapterListFor(comic *Comic) { //TODO: skipAllowed boolean (optimisation, download only last page to update existing list, the suggestion may be disregarded) - only some plugins
	for _, source := range comic.Sources() {
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
	}
}

func (this *fetcher) DownloadPageLinksFor(comic *Comic, chapterIndex, scanlationIndex int) (success bool) {
	chapter, identity := comic.GetChapter(chapterIndex)
	scanlation := chapter.Scanlation(scanlationIndex)
	if plugin, success := this.plugins[scanlation.PluginName]; success && plugin.Capabilities().ProvidesData {
		links := plugin.fetchChapterPageLinks(scanlation.URL)
		scanlation.PageLinks = links
		chapter.AddScanlation(scanlation)    //reinsert after modifying
		comic.AddChapter(identity, &chapter) //reinsert //TODO: use pointers instead?
	}
	return
}

func (this *fetcher) PluginNameFromURL(url string) (FetcherPluginName, error) {
	for pluginName, plugin := range this.plugins {
		if plugin.IsURLValid(url) {
			return pluginName, nil
		}
	}
	return "", errors.New("Plugin autodetect failed!")
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
