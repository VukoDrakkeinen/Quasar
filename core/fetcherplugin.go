package core

import (
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"strconv"
)

var (
	shared_rRemoveHTML = qregexp.MustCompile(`<[^>]+>`)
)

type FetcherPluginName string

type FetcherPlugin interface { //TODO: shared implementation
	PluginName() FetcherPluginName
	HumanReadableName() string
	Languages() []string
	Capabilities() FetcherPluginCapabilities
	Settings() PerPluginSettings
	SetSettings(new PerPluginSettings)
	IsURLValid(url string) bool
	findComic(title, author string, genres []idsdict.ComicGenreId, status comicStatus, ctype comicType, mature bool) []comicSearchResult
	fetchAdvert() advert
	fetchComicInfo(comic *Comic) *ComicInfo
	fetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool)
	fetchChapterPageLinks(url string) []string
	setFetcher(parent *fetcher)
	findComicURL(title string) string //TODO: remove
}

type FetcherPluginCapabilities struct {
	ProvidesMetadata bool
	ProvidesData     bool
}

type advert struct {
	simple   bool
	imageURL string
	link     string
	html     []byte
}

type fetcherPluginSharedImpl struct {
	name      FetcherPluginName
	settings  PerPluginSettings
	m_fetcher *fetcher
}

func (this *fetcherPluginSharedImpl) fetcher() *fetcher {
	if this.m_fetcher == nil {
		panic("Attempted to use orphaned plugin " + this.name + "!")
	}
	return this.m_fetcher
}

func (this *fetcherPluginSharedImpl) setFetcher(parent *fetcher) {
	this.m_fetcher = parent
}

func (this *fetcherPluginSharedImpl) PluginName() FetcherPluginName {
	return this.name
}

func (this *fetcherPluginSharedImpl) Settings() PerPluginSettings {
	return this.settings
}

func (this *fetcherPluginSharedImpl) SetSettings(new PerPluginSettings) {
	var maxConns uint
	if overrideMaxConns := new.OverrideDefaults[4]; overrideMaxConns {
		maxConns = new.MaxConnectionsToHost
	}
	this.fetcher().PluginLimitsUpdated(this.name, maxConns)
	this.settings = new
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
