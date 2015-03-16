package redshift

import "strconv"

type FetcherPluginName string
type FetcherPlugin interface {
	PluginName() FetcherPluginName
	Languages() []string
	Capabilities() FetcherPluginCapabilities
	IsURLValid(url string) bool
	findComicURL(title string) string
	findComicURLList(title string) (links []string, titles []string)
	fetchComicInfo(comic *Comic) *ComicInfo
	fetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool)
	fetchChapterPageLinks(url string) []string
	setFetcher(parent *fetcher)
}

type FetcherPluginCapabilities struct { //TODO: more detailed capabilities?
	ProvidesInfo bool
	ProvidesData bool
}

func titleFromIdentity(identity ChapterIdentity) string {
	title := "[Chapter #" + strconv.FormatInt(int64(identity.MajorNum), 10)
	if identity.MinorNum != 0 {
		title += "." + strconv.FormatInt(int64(identity.MinorNum), 10)
	}
	title += "]"
	return title
}
