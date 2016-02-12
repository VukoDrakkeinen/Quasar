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

type SourceId string

type Source interface {
	Id() SourceId
	Name() string
	Languages() []string
	Capabilities() SourceCapabilities
	Config() SourceConfig
	SetConfig(cfg SourceConfig)
	IsURLValid(url string) bool
	search(title, author string, genres []idsdict.ComicGenreId, status comicStatus, ctype comicType, mature bool) []comicSearchResult
	advert() advert
	comicInfo(comic *Comic) *ComicInfo
	chapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool)
	chapterDataLinks(url string) []string
	setParent(parent *fetcher)
	comicURL(title string) string //TODO: remove
}

type SourceCapabilities struct {
	ProvidesMetadata bool
	ProvidesData     bool
}

type advert struct {
	simple   bool
	imageURL string
	link     string
	html     []byte
}

type sourceSharedImpl struct {
	id        SourceId
	settings  SourceConfig
	m_fetcher *fetcher
}

func (this *sourceSharedImpl) fetcher() *fetcher {
	if this.m_fetcher == nil {
		panic("Attempted to use orphaned plugin " + this.id + "!")
	}
	return this.m_fetcher
}

func (this *sourceSharedImpl) setParent(parent *fetcher) {
	this.m_fetcher = parent
}

func (this *sourceSharedImpl) Id() SourceId {
	return this.id
}

func (this *sourceSharedImpl) Config() SourceConfig {
	return this.settings
}

func (this *sourceSharedImpl) SetConfig(cfg SourceConfig) {
	var maxConns uint
	if overrideMaxConns := cfg.OverrideDefaults[4]; overrideMaxConns {
		maxConns = cfg.MaxConnectionsToHost
	}
	this.fetcher().PluginLimitsUpdated(this.id, maxConns)
	this.settings = cfg
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

func (this *SourceId) Scan(src interface{}) error {
	switch s := src.(type) {
	case string: //yeah, can't do "case string, []byte", can't fallthrough for some reason. o_O Google fix pls
		*this = SourceId(s)
		return nil
	case []byte:
		*this = SourceId(s)
		return nil
	default:
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be a string or []uint8, got %T!)", *this, src))
	}
}
