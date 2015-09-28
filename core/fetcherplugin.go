package core

import (
	"errors"
	"fmt"
	"strconv"
)

type FetcherPluginName string

//TODO: also fetch adverts (we don't want to leech!)
type FetcherPlugin interface { //TODO: shared implementation
	PluginName() FetcherPluginName
	HumanReadableName() string
	Languages() []string
	Capabilities() FetcherPluginCapabilities
	Settings() PerPluginSettings
	SetSettings(new PerPluginSettings)
	IsURLValid(url string) bool
	findComicURL(title string) string
	findComicURLList(title string) (links []string, titles []string) //TODO: proper search func
	fetchComicInfo(comic *Comic) *ComicInfo
	fetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool)
	fetchChapterPageLinks(url string) []string
	setFetcher(parent *fetcher)
}

type FetcherPluginCapabilities struct { //TODO: more detailed capabilities?
	ProvidesMetadata bool
	ProvidesData     bool
}

func titleFromIdentity(identity ChapterIdentity) string {
	title := "[Chapter #" + strconv.FormatInt(int64(identity.MajorNum), 10)
	if identity.MinorNum != 0 {
		title += "." + strconv.FormatInt(int64(identity.MinorNum), 10)
	}
	if identity.Letter != 0 {
		title += strconv.FormatInt(int64(identity.Letter+'a'-1), 10)
	}
	title += "]"
	return title
}

func (this *FetcherPluginName) Scan(src interface{}) error {
	switch s := src.(type) {
	case string: //yeah, can't do "case string, []byte", can't fallthrough for some reason. o_O Google fix pls
		*this = FetcherPluginName(s)
		return nil
	case []byte:
		*this = FetcherPluginName(s)
		return nil
	default:
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be a string or []uint8, got %T!)", *this, src))
	}
}
