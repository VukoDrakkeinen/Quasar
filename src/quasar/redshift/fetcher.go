package redshift

import (
	"errors"
	"io/ioutil"
	"net/http"
	. "quasar/redshift/idbase"
)

type FetcherPluginName string
type FetcherPlugin interface {
	SetFetcher(parent *Fetcher)
	PluginName() FetcherPluginName
	Languages() []string
	Capabilities() FetcherPluginCapabilities
	IsURLValid(url string) bool
	FindComicURL(title string) string
	FindComicURLList(title string) (links []string, titles []string)
	FetchComicInfo(comic *Comic) *ComicInfo
	FetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter)
	FetchChapterPageLinks(comic *Comic, chapterIndex, alterIndex int) []string
}

type FetcherPluginCapabilities struct { //TODO: more detailed capabilities?
	ProvidesInfo bool
	ProvidesData bool
}

type Fetcher struct { //TODO: handle missing plugin errors gracefully
	plugins   map[FetcherPluginName]FetcherPlugin
	webClient *http.Client
	settings  *GlobalSettings
}

func (this *Fetcher) initialize() {
	if this.plugins == nil {
		this.plugins = make(map[FetcherPluginName]FetcherPlugin)
		this.webClient = &http.Client{
			CheckRedirect: nil, //TODO: write the redirect handling function
		}
		this.settings = NewGlobalSettings()
	}
}

func (this *Fetcher) RegisterPlugin(plugin FetcherPlugin) (success, replaced bool) {
	this.initialize()
	name := plugin.PluginName()
	oldPlugin, replaced := this.plugins[name]
	if replaced {
		oldPlugin.SetFetcher(nil)
	}
	this.plugins[name] = plugin
	plugin.SetFetcher(this)
	LangDict.AssignIds(plugin.Languages())
	success = true //TODO?
	return
}

func (this *Fetcher) DownloadComicInfoFor(comic *Comic) {
	this.initialize()
	for _, source := range comic.Sources() {
		comic.Info.Merge(this.plugins[source.PluginName].FetchComicInfo(comic))
	}
}

//TODO: proper error handling
//TODO: parallelization?
func (this *Fetcher) DownloadData(url string) []byte {
	this.initialize()
	response, err := this.webClient.Get(url)
	if err != nil {
		panic("Error in Client.Get()")
	}
	defer response.Body.Close()

	data, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic("Error in ioutil.ReadAll()")
	}
	return data
}

/*
//TODO: proper error handling
func (this *Fetcher) DownloadPageImage(index int, chapter comic.Chapter /*, com comic.Comic) image.Image {
	request, err := http.NewRequest("GET", chapter.Data[pluginName].PageLinks[index], nil)
	if err != nil {
		fmt.Println("NewRequest error:", err)
		return nil
	}

	response, err := this.theHTTPClient.Do(request)
	if err != nil {
		fmt.Println("ClientDo error:", err)
		return nil
	}

	if sc := response.StatusCode; sc != 200 {
		fmt.Println("Status Code invalid:", sc, "!")
		return nil
	}

	reader := response.Body
	if response.ContentLength == 0 {
		fmt.Println("LENGTH 0")
		return nil
	}

	binaryData, err := jpeg.Decode(reader)
	if err != nil {
		fmt.Println("Decoding error:", err)
	}

		//binaryData, err := ioutil.ReadAll(reader)
		//if err != nil {
		//	fmt.Println("ReadAll error:", err)
		//}
	return binaryData
}//*/

func (this *Fetcher) DownloadChapterListFor(comic *Comic) { //TODO: whole bool (optimisation, download only last page to update existing list) - only some plugins
	this.initialize()
	for _, source := range comic.Sources() {
		identities, chapters := this.plugins[source.PluginName].FetchChapterList(comic)
		comic.AddMultipleChapters(identities, chapters)
	}
}

func (this *Fetcher) DownloadPageLinksFor(comic *Comic, chapterIndex, alterIndex int) {
	this.initialize()
	for _, source := range comic.Sources() {
		if plugin := this.plugins[source.PluginName]; plugin.Capabilities().ProvidesData {
			links := plugin.FetchChapterPageLinks(comic, chapterIndex, alterIndex)
			chapter, identity := comic.GetChapter(chapterIndex)
			data := chapter.DataForPlugin(source.PluginName, alterIndex)
			data.PageLinks = links
			chapter.SetData(source.PluginName, data)
			comic.AddChapter(identity, &chapter)
		}
	}
}

func (this *Fetcher) PluginNameFromURL(url string) (FetcherPluginName, error) {
	this.initialize()
	for pluginName, plugin := range this.plugins {
		if plugin.IsURLValid(url) {
			return pluginName, nil
		}
	}
	return "", errors.New("Plugin autodetect failed!")
}

func (this *Fetcher) TestFind(comic *Comic, pluginName FetcherPluginName, comicTitle string) {
	this.initialize()
	plugin := this.plugins[pluginName]
	urlFound := plugin.FindComicURL(comicTitle)
	if urlFound != "" {
		comic.AddSource(UpdateSource{
			PluginName: pluginName,
			URL:        urlFound,
			MarkAsRead: false,
		})
	}
}
