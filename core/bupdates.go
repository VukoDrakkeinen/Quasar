package core

import (
	"bytes"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"html"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

var (
	bakaUpdates_rURLValidator = qregexp.MustCompile(`^https?://www.mangaupdates.com/series.html\?id=\d+$`)

	bakaUpdates_rURLAndTitleList = qregexp.MustCompile(`(https?://www.mangaupdates.com/series.html\?id=\d+)' alt='Series Info'>(?:<i>)?([^<]+)`)

	bakaUpdates_rInfoRegion  = qregexp.MustCompile(`(?s)<!-- Start:Series Info-->.*<!-- End:Series Info-->`) //woot, useful comments in code!
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
	//bakaUpdates_rAuthorsAndArtists = qregexp.MustCompile(`(?<=title='Author Info'><u>).*(?=</u)`)

	bakaUpdates_rChaptersRegion   = qregexp.MustCompile(`(?s)<!-- Start:Center Content -->.*<!-- End:Center Content -->`)
	bakaUpdates_rChpListPageCount = qregexp.MustCompile(`(?<=nowrap>Pages \()\d+(?=\) <a href=)`)
	bakaUpdates_rChaptersInfoList = qregexp.MustCompile(`(?<=<tr >\r?\n)(?:\s+<td class='text pad.*</td>\r?\n){5}(?=\s+</tr>)`)
	bakaUpdates_rChpInfoPieceList = qregexp.MustCompile(`(?<=' ?>)([^<]*)(?=</(?:a|td)>)`)

	bakaUpdates_rIdentityParse = qregexp.MustCompile(`^(\d+)(?:-(\d+))?(?:\.(\d))?(?: ([LH]Q))?(?: ?\(?v(\d)\)?)?( Color)?( \+ Extra)?`)

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
		ProvidesInfo: true,
		ProvidesData: false,
	}
}

func (this *bakaUpdates) Settings() PerPluginSettings {
	return this.settings
}

func (this *bakaUpdates) SetSettings(new PerPluginSettings) {
	if overrideMaxConns := new.OverrideDefaults[4]; overrideMaxConns {
		this.fetcher().connLimits[this.name] = new.MaxConnectionsToHost
	} else {
		this.fetcher().connLimits[this.name] = 0
	}
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
	title := string(bakaUpdates_rTitle.Find(infoRegion))
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
	if imageUrl != "" {
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
	source := comic.GetSource(this.name)
	linkPrefix := "https://www.mangaupdates.com/releases.html?stype=series&perpage=100&search=" + bakaUpdates_rComicID.FindString(source.URL) + "&page="
	region, err := this.fetcher().DownloadData(this.name, linkPrefix+"1", false)
	if err != nil {
		panic(err)
	}
	firstRegion := bakaUpdates_rChaptersRegion.Find(region)
	pageCountString := string(bakaUpdates_rChpListPageCount.Find(firstRegion))
	pageCount, _ := strconv.ParseUint(pageCountString, 10, 32)
	regionsSlice := make([][]byte, pageCount)
	regionsSlice[0] = firstRegion
	var wg sync.WaitGroup
	for i := 2; i <= int(pageCount); i++ {
		i := i
		wg.Add(1)
		go func() {
			region, err := this.fetcher().DownloadData(this.name, linkPrefix+strconv.FormatInt(int64(i), 10), false)
			if err != nil {
				panic(err)
			}
			regionsSlice[i-1] = bakaUpdates_rChaptersRegion.Find(region)
			wg.Done()
		}()
	}
	wg.Wait()
	identities = make([]ChapterIdentity, 0, pageCount*100)
	chapters = make([]Chapter, 0, pageCount*100)

	for i := len(regionsSlice) - 1; i >= 0; i-- { //cannot use range, because we're iterating in reverse :(
		chaptersInfos := bakaUpdates_rChaptersInfoList.FindAll(regionsSlice[i], -1)
		prevIdentity := ChapterIdentity{}
		for j := len(chaptersInfos) - 1; j >= 0; j-- { //I really wish they added revrange keyword
			infoPieces := bakaUpdates_rChpInfoPieceList.FindAll(chaptersInfos[j], -1)
			// [0] is date
			// [1] is comic title (wut)
			// [2] is volume number
			// [3] is chapter number/s
			// [4-?] is scanlators
			volumeString := string(infoPieces[2])
			missingVolumes = missingVolumes || volumeString == ""
			numberString := string(infoPieces[3])
			scanlatorNames := infoPieces[4:]
			for i, scanlator := range scanlatorNames {
				scanlatorNames[i] = []byte(html.UnescapeString(string(scanlator)))
			}
			scanlators, _ := Scanlators.AssignIdsBytes(scanlatorNames)
			newIdentities, err := parseBakaIdentities(volumeString, numberString, prevIdentity)
			if err != nil {
				qlog.Logf(qlog.Error, "Parsing identity \"%s\"+\"%s\" failed: %v\n", volumeString, numberString, err)
			}
			prevIdentity = newIdentities[len(newIdentities)-1]
			for _, identity := range newIdentities {
				chapter := NewChapter(source.MarkAsRead)
				chapter.AddScanlation(ChapterScanlation{
					Title:      titleFromIdentity(identity),
					Language:   ENGLISH_LANG(),
					Scanlators: JoinScanlators(scanlators),
					PluginName: this.name,
					URL:        "",
					PageLinks:  make([]string, 0),
				})

				identities = append(identities, identity)
				chapters = append(chapters, *chapter)
			}
		}
	}
	return
}

func (this *bakaUpdates) fetchChapterPageLinks(url string) []string {
	_ = url           //unused
	return []string{} //plugin doesn't provide data, return empty list
}

func parseBakaIdentities(volumeStr, numberStr string, previous ChapterIdentity) (identities []ChapterIdentity, err error) {
	inputStr := strconv.Quote(volumeStr) + " + " + strconv.Quote(numberStr)
	qualityModifier := MQ_Modifier
	identity := ChapterIdentity{Version: qualityModifier + 1}
	volumeParsing := bakaUpdates_rIdentityParse.FindStringSubmatch(volumeStr)
	/*
		[0] is entire match
		[1] is volume number (starting)
		[2] is volume number (ending, optional; we'll have to do some guessing with chapters per volume if present)
		[3] is unused
		[4] is LQ/HQ marker (optional)
		[5] is version number (optional)
		[6] is Color marker (optional, set chapter minor number to 1)
		[7] is unused
	*/
	var lastVolume byte
	if len(volumeParsing) > 0 {
		if str := volumeParsing[1]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Volume = byte(i)
			lastVolume = byte(i)
		}
		if str := volumeParsing[2]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			lastVolume = byte(i)
		}
		if str := volumeParsing[4]; str != "" {
			if str == "LQ" {
				qualityModifier = LQ_Modifier
			} else if str == "HQ" {
				qualityModifier = HQ_Modifier
			}
		}
		if str := volumeParsing[5]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Version = qualityModifier + byte(i)
		}
		if str := volumeParsing[6]; str != "" {
			identity.MinorNum = 1
		}
	}

	numberParsing := bakaUpdates_rIdentityParse.FindStringSubmatch(numberStr)
	/*
		[0] is entire match
		[1] is starting chapter major number
		[2] is ending chapter major number (optional)
		[3] is starting chapter minor number (optional)
		[4] is LQ/HQ marker (optional)
		[5] is chapters version (optional)
		[6] is Color marker (optional, add +1 to chapter minor number)
		[7] is Extra Chapter marker (optional)
	*/
	if len(numberParsing) > 0 {
		/*        numberParsing[1]         */ {
			i, _ := strconv.ParseUint(numberParsing[1], 10, 16)
			identity.MajorNum = uint16(i)
		}
		if str := numberParsing[3]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.MinorNum = byte(i)
		}
		if str := numberParsing[4]; str != "" {
			if str == "LQ" {
				qualityModifier = LQ_Modifier
			} else if str == "HQ" {
				qualityModifier = HQ_Modifier
			}
		}
		if str := numberParsing[5]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Version = qualityModifier + byte(i)
		}
		if str := numberParsing[6]; str != "" {
			identity.MinorNum += 1
		}
		identities = append(identities, identity)
		lastChapter := identity.MajorNum
		if str := numberParsing[2]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 16)
			lastChapter = uint16(i)
			volumesCount := (byte(lastVolume) - identity.Volume + 1)
			chaptersPerVolume := (byte(lastChapter)-byte(identity.MajorNum))/volumesCount + 1 //yep, we're guessing
			for j := identity.MajorNum + 1; j <= lastChapter; j++ {
				volPlus := byte((j - identity.MajorNum)) / (chaptersPerVolume)
				identities = append(identities, ChapterIdentity{identity.Volume + volPlus, j, identity.MinorNum, identity.Letter, identity.Version})
			}
		}
		if numberParsing[7] != "" {
			identities = append(identities, ChapterIdentity{lastVolume, lastChapter, identity.MinorNum + 1, identity.Letter, identity.Version})
		}
		return
	} else if numberStr != "" { //Apparently bullshit like "Road to the movie" is a valid chapter number
		identity.MajorNum = previous.MajorNum
		identity.MinorNum += 1 //so we treat it as a special chapter
		identities = append(identities, identity)
		return
	} else { //numberStr is empty, which means whole volume got scanlated, but we have no way to tell how many chapters is that
		return []ChapterIdentity{}, qerr.NewParse("Whole volume scanlated, unknown number of chapters", nil, inputStr)
	}
}
