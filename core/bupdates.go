package core

import (
	"bytes"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"html"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
)

var (
	bakaUpdates_rURLValidator = qregexp.MustCompile(`^https?://www.mangaupdates.com/series.html\?id=\d+$`)

	bakaUpdates_rURLAndTitleList = qregexp.MustCompile(`(https?://www.mangaupdates.com/series.html\?id=\d+)' alt='Series Info'>(?:<i>)?([^<]+)`)

	bakaUpdates_rInfoRegion  = qregexp.MustCompile(`(?s)<!-- Start:Series Info-->.*<!-- End:Series Info-->`) //what is this insanity, useful comments in the page code?!
	bakaUpdates_rTitle       = qregexp.MustCompile(`(?<=tabletitle">)[^<]+`)
	bakaUpdates_rDescription = qregexp.MustCompile(`(?<=div class="sCat"><b>Description</b></div>\n<div class="sContent" style="text-align:justify">).*`)
	bakaUpdates_rRemoveHTML  = qregexp.MustCompile(`<[^>]+>`)
	bakaUpdates_rType        = qregexp.MustCompile(`(?<=<div class="sCat"><b>Type</b></div>\n<div class="sContent" >).*`)
	bakaUpdates_rAltTitles   = qregexp.MustCompile(`(?<=<div class="sCat"><b>Associated Names</b></div>\n<div class="sContent" >).*(?=<)`)
	bakaUpdates_rStatus      = qregexp.MustCompile(`(?<=<div class="sCat"><b>Status in Country of Origin</b></div>\n<div class="sContent" >)[^(]+\(([^)]+)\)`)
	bakaUpdates_rScanStatus  = qregexp.MustCompile(`(?<=<div class="sCat"><b>Completely Scanlated\?</b></div>\n<div class="sContent" >).*`)
	bakaUpdates_rRating      = qregexp.MustCompile(`(?<=<div class="sCat"><b>User Rating</b></div>\n<div class="sContent" >Average:)[^<]+<br>Bayesian Average: <b>(\d{1,2}\.\d\d)`)
	bakaUpdates_rImageURL    = qregexp.MustCompile(`https?://www.mangaupdates.com/image/[^']+`)
	bakaUpdates_rGenres      = qregexp.MustCompile(`(?<=<div class="sCat"><b>Genre</b></div>\n<div class="sContent" >).*(?=&)`)
	bakaUpdates_rCategories  = qregexp.MustCompile(`(?<=,\d\)'>).*(?=</a></li>)`)
	bakaUpdates_rAuthorsLine = qregexp.MustCompile(`(?<=<div class="sCat"><b>Author\(s\)</b></div>\n<div class="sContent" >).*`)
	bakaUpdates_rArtistsLine = qregexp.MustCompile(`(?<=<div class="sCat"><b>Artist\(s\)</b></div>\n<div class="sContent" >).*`)
	bakaUpdates_rExtract     = qregexp.MustCompile(`(?<=<u>)[^<]+(?=</u>)`)

	bakaUpdates_rComicID = qregexp.MustCompile(`(?<=id=)\d+`) //TODO: UpdateSource holding additional plugin data?
)

type bakaUpdates struct {
	name      FetcherPluginName
	settings  PerPluginSettings
	m_fetcher *fetcher
}

func NewBakaUpdates() *bakaUpdates { //TODO: logic saved as interpreted files
	ret := &bakaUpdates{}
	ret.name = FetcherPluginName(reflect.TypeOf(*ret).Name())
	return ret
}

func (this *bakaUpdates) fetcher() *fetcher { //TODO: don't panic, just log
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	return this.m_fetcher
}

func (this *bakaUpdates) setFetcher(parent *fetcher) {
	this.m_fetcher = parent
}

func (this *bakaUpdates) PluginName() FetcherPluginName {
	return this.name
}

func (this *bakaUpdates) HumanReadableName() string {
	return "Baka-Updates"
}

func (this *bakaUpdates) Languages() []string {
	return []string{"English"}
}

func (this *bakaUpdates) Capabilities() FetcherPluginCapabilities {
	return FetcherPluginCapabilities{
		ProvidesMetadata: false,
		ProvidesData:     false,
	}
}

func (this *bakaUpdates) Settings() PerPluginSettings {
	return this.settings
}

func (this *bakaUpdates) SetSettings(new PerPluginSettings) {
	var maxConns uint
	if overrideMaxConns := new.OverrideDefaults[4]; overrideMaxConns {
		maxConns = new.MaxConnectionsToHost
	}
	this.fetcher().PluginLimitsUpdated(this.name, maxConns)
	this.settings = new
}

func (this *bakaUpdates) IsURLValid(url string) bool {
	return bakaUpdates_rURLValidator.MatchString(url)
}

func (this *bakaUpdates) findComicURL(title string) string {
	links, titles := this.findComicURLList(title)
	for i, ctitle := range titles {
		if strings.EqualFold(title, ctitle) {
			return links[i]
		}
	}
	return ""
}

func (this *bakaUpdates) findComicURLList(title string) (links []string, titles []string) { //TODO: go trough ALL the subpages
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	contents, err := this.fetcher().DownloadData(
		this.name,
		"https://www.mangaupdates.com/series.html?page=1&stype=title&perpage=100&search="+url.QueryEscape(title),
		false,
	)
	if err != nil {
		panic(err)
	}
	urlAndTitleList := bakaUpdates_rURLAndTitleList.FindAllSubmatch(contents, -1)
	for _, urlAndTitle := range urlAndTitleList {
		links = append(links, string(urlAndTitle[1]))
		titles = append(titles, html.UnescapeString(string(urlAndTitle[2])))
	}
	return
}

func (this *bakaUpdates) fetchComicInfo(comic *Comic) *ComicInfo {
	url := comic.GetSource(this.name).URL
	contents, err := this.fetcher().DownloadData(this.name, url, false)
	if err != nil {
		panic(err)
	}
	infoRegion := bakaUpdates_rInfoRegion.Find(contents)
	title := html.UnescapeString(string(bakaUpdates_rTitle.Find(infoRegion)))
	description := html.UnescapeString(string(bakaUpdates_rRemoveHTML.ReplaceAll(
		bytes.Replace(bakaUpdates_rDescription.Find(infoRegion), []byte("<BR>"), []byte("\n"), -1),
		[]byte{},
	)))
	cType := InvalidComic
	switch string(bakaUpdates_rType.Find(infoRegion)) {
	case "Manga":
		cType = Manga
	case "Manhwa":
		cType = Manhwa
	case "Manhua":
		cType = Manhua
	default:
		cType = Other
	}
	altTitles := make(map[string]struct{})
	for _, altTitle := range bytes.Split(bakaUpdates_rAltTitles.Find(infoRegion), []byte("<br />")) {
		altTitles[html.UnescapeString(string(altTitle))] = struct{}{}
	}
	statusString := string(bakaUpdates_rStatus.Find(infoRegion))
	status := ComicStatusInvalid
	switch {
	case statusString == "Ongoing":
		status = ComicOngoing
	case statusString == "Complete":
		status = ComicComplete
	case statusString == "Hiatus":
		status = ComicOnHiatus
	case statusString == "Complete/Discontinued":
		status = ComicDiscontinued
	}
	scanStatus := ScanlationStatusInvalid
	switch string(bakaUpdates_rScanStatus.Find(infoRegion)) {
	case "Yes":
		scanStatus = ScanlationComplete
	case "No":
		scanStatus = ScanlationOngoing
	}
	ratingString := string(bakaUpdates_rRating.Find(infoRegion))
	rating, _ := strconv.ParseFloat(ratingString, 32)
	var thumbnailFilename string
	imageUrl := string(bakaUpdates_rImageURL.Find(infoRegion))
	if imageUrl != "" { //TODO
		thumbnailFilename = path.Base(imageUrl)
		thumbnail, err := this.fetcher().DownloadData(this.name, imageUrl, false)
		if err != nil {
			panic(err)
		}
		qdb.SaveThumbnail(thumbnailFilename, thumbnail)
	}
	genres := make(map[ComicGenreId]struct{})
	for _, genre := range qutils.Vals(ComicGenres.AssignIdsBytes(bytes.Split(bakaUpdates_rRemoveHTML.ReplaceAll(bakaUpdates_rGenres.Find(infoRegion), []byte{}), []byte("&nbsp; "))))[0].([]ComicGenreId) {
		genres[genre] = struct{}{}
	}
	ajax, err := this.fetcher().DownloadData(
		this.name, "https://www.mangaupdates.com/ajax/show_categories.php?type=1&s="+bakaUpdates_rComicID.FindString(url),
		false,
	)
	if err != nil {
		panic(err)
	}
	categories := make(map[ComicTagId]struct{})
	for _, tag := range qutils.Vals(ComicTags.AssignIdsBytes(bakaUpdates_rCategories.FindAll(ajax, -1)))[0].([]ComicTagId) {
		categories[tag] = struct{}{}
	}
	authors, _ := Authors.AssignIdsBytes(bakaUpdates_rExtract.FindAll(bakaUpdates_rAuthorsLine.Find(infoRegion), -1))
	artists, _ := Artists.AssignIdsBytes(bakaUpdates_rExtract.FindAll(bakaUpdates_rArtistsLine.Find(infoRegion), -1))
	_, mature := genres[MATURE_GENRE()]

	return &ComicInfo{
		Title:             title,
		AltTitles:         altTitles,
		Authors:           authors,
		Artists:           artists,
		Genres:            genres,
		Categories:        categories,
		Type:              cType,
		Status:            status,
		ScanlationStatus:  scanStatus,
		Description:       description,
		Rating:            float32(rating),
		Mature:            mature,
		ThumbnailFilename: thumbnailFilename,
	}
}

func (this *bakaUpdates) fetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool) {
	_ = comic //unused
	return    //plugin doesn't provide metadata, return empty lists
}

func (this *bakaUpdates) fetchChapterPageLinks(url string) []string {
	_ = url           //unused
	return []string{} //plugin doesn't provide data, return empty list
}
