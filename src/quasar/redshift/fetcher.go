package redshift

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	. "quasar/redshift/idsdict"
	"sort"
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
}

func NewFetcher(settings *GlobalSettings, plugins ...FetcherPlugin) *fetcher {
	fet := &fetcher{
		plugins: make(map[FetcherPluginName]FetcherPlugin),
		webClient: &http.Client{
			CheckRedirect: nil, //TODO: write the redirect handling function
		},
		settings: settings,
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

//TODO: proper error handling
//TODO: parallelization?
func (this *fetcher) DownloadData(url string) []byte {
	response, err := this.webClient.Get(url)
	fmt.Println("Response status:", response.Status)
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
