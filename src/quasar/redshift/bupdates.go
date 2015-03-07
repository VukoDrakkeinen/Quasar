package redshift

import (
	"bytes"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"quasar/qregexp"
	"quasar/qutils"
	. "quasar/redshift/idbase"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type BUpdates struct {
	initialized bool
	name        FetcherPluginName
	m_fetcher   *Fetcher

	rURLValidator *qregexp.QRegexp

	rURLAndTitleList *qregexp.QRegexp

	rInfoRegion  *qregexp.QRegexp
	rTitle       *qregexp.QRegexp
	rDescription *qregexp.QRegexp
	rRemoveHTML  *qregexp.QRegexp
	rType        *qregexp.QRegexp
	rAltTitles   *qregexp.QRegexp
	rStatus      *qregexp.QRegexp
	rScanStatus  *qregexp.QRegexp
	rRating      *qregexp.QRegexp
	rImageURL    *qregexp.QRegexp
	rGenres      *qregexp.QRegexp
	rCategories  *qregexp.QRegexp
	rAuthorsLine *qregexp.QRegexp
	rArtistsLine *qregexp.QRegexp
	rExtract     *qregexp.QRegexp
	//rAuthorsAndArtists *qregexp.QRegexp

	rChaptersRegion   *qregexp.QRegexp
	rChpListPageCount *qregexp.QRegexp
	rChaptersInfoList *qregexp.QRegexp
	rChpInfoPieceList *qregexp.QRegexp

	rIdentityParse *qregexp.QRegexp

	rComicID *qregexp.QRegexp
}

func NewBUpdates() *BUpdates {
	return new(BUpdates).initialize()
}

func (this *BUpdates) initialize() *BUpdates { //TODO: handle errors
	if !this.initialized { //TODO: logic saved as interpreted files
		this.name = FetcherPluginName(reflect.TypeOf(*this).Name())

		this.rURLValidator = qregexp.MustCompile(`^http://www.mangaupdates.com/series.html\?id=\d+$`)

		this.rURLAndTitleList = qregexp.MustCompile(`(https://www.mangaupdates.com/series.html\?id=\d+)' alt='Series Info'>(?:<i>)?([^<]+)`)

		this.rInfoRegion = qregexp.MustCompile(`(?s)<!-- Start:Series Info-->.*<!-- End:Series Info-->`) //woot, useful comments in code!
		this.rTitle = qregexp.MustCompile(`(?<=tabletitle">)[^<]+`)
		this.rDescription = qregexp.MustCompile(`(?<=div class="sCat"><b>Description</b></div>\n<div class="sContent" style="text-align:justify">).*`)
		this.rRemoveHTML = qregexp.MustCompile(`<[^>]+>`)
		this.rType = qregexp.MustCompile(`(?<=<div class="sCat"><b>Type</b></div>\n<div class="sContent" >).*`)
		this.rAltTitles = qregexp.MustCompile(`(?<=<div class="sCat"><b>Associated Names</b></div>\n<div class="sContent" >).*(?=<)`)
		this.rStatus = qregexp.MustCompile(`(?<=<div class="sCat"><b>Status in Country of Origin</b></div>\n<div class="sContent" >)[^(]+\(([^)]+)\)`)
		this.rScanStatus = qregexp.MustCompile(`(?<=<div class="sCat"><b>Completely Scanlated\?</b></div>\n<div class="sContent" >).*`)
		this.rRating = qregexp.MustCompile(`(?<=<div class="sCat"><b>User Rating</b></div>\n<div class="sContent" >Average:)[^<]+<br>Bayesian Average: <b>(\d{1,2}\.\d\d)`)
		this.rImageURL = qregexp.MustCompile(`https://www.mangaupdates.com/image/[^']+`)
		this.rGenres = qregexp.MustCompile(`(?<=<div class="sCat"><b>Genre</b></div>\n<div class="sContent" >).*(?=&)`)
		this.rCategories = qregexp.MustCompile(`(?<=,\d\)'>).*(?=</a></li>)`)
		this.rAuthorsLine = qregexp.MustCompile(`(?<=<div class="sCat"><b>Author\(s\)</b></div>\n<div class="sContent" >).*`)
		this.rArtistsLine = qregexp.MustCompile(`(?<=<div class="sCat"><b>Artist\(s\)</b></div>\n<div class="sContent" >).*`)
		this.rExtract = qregexp.MustCompile(`(?<=<u>)[^<]+(?=</u>)`)
		//this.rAuthorsAndArtists = qregexp.MustCompile(`(?<=title='Author Info'><u>).*(?=</u)`)

		this.rChaptersRegion = qregexp.MustCompile(`(?s)<!-- Start:Center Content -->.*<!-- End:Center Content -->`)
		this.rChpListPageCount = qregexp.MustCompile(`(?<=nowrap>Pages \()\d+(?=\) <a href=)`)
		this.rChaptersInfoList = qregexp.MustCompile(`(?<=<tr >\r?\n)(?:\s+<td class='text pad.*</td>\r?\n){5}(?=\s+</tr>)`)
		this.rChpInfoPieceList = qregexp.MustCompile(`(?<=' ?>)([^<]*)(?=</(?:a|td)>)`)

		this.rIdentityParse = qregexp.MustCompile(`^(\d+)(?:-(\d+))?(?:\.(\d))?(?: ([LH]Q))?(?: ?\(?v(\d)\)?)?( Color)?( \+ Extra)?`)

		this.rComicID = qregexp.MustCompile(`(?<=id=)\d+`) //FIXME: UpdateSource holding additional plugin data?

		this.initialized = true
		fmt.Println("Plugin", this.name, "initialized!")
	}
	return this
}

func (this *BUpdates) fetcher() *Fetcher { //TODO: don't panic, just log
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	return this.m_fetcher
}

func (this *BUpdates) SetFetcher(parent *Fetcher) {
	this.initialize()
	this.m_fetcher = parent
}

func (this *BUpdates) PluginName() FetcherPluginName {
	this.initialize()
	return this.name
}

func (this *BUpdates) Languages() []string {
	return []string{"English"}
}

func (this *BUpdates) Capabilities() FetcherPluginCapabilities {
	this.initialize()
	return FetcherPluginCapabilities{
		ProvidesInfo: true,
		ProvidesData: false,
	}
}

func (this *BUpdates) IsURLValid(url string) bool {
	return this.rURLValidator.MatchString(url)
}

func (this *BUpdates) FindComicURL(title string) string {
	this.initialize()
	links, titles := this.FindComicURLList(title)
	for i, ctitle := range titles {
		if strings.EqualFold(title, ctitle) {
			return links[i]
		}
	}
	return ""
}

func (this *BUpdates) FindComicURLList(title string) (links []string, titles []string) { //TODO: go trough ALL the subpages
	this.initialize()
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	contents := this.fetcher().DownloadData("https://www.mangaupdates.com/series.html?page=1&stype=title&perpage=100&search=" + title)
	urlAndTitleList := this.rURLAndTitleList.FindAllSubmatch(contents, -1)
	for _, urlAndTitle := range urlAndTitleList {
		links = append(links, string(urlAndTitle[1]))
		titles = append(titles, html.UnescapeString(string(urlAndTitle[2])))
	}
	return
}

func (this *BUpdates) FetchComicInfo(comic *Comic) *ComicInfo {
	this.initialize()
	url := comic.GetSource(this.name).URL
	contents := this.fetcher().DownloadData(url)
	infoRegion := this.rInfoRegion.Find(contents)
	title := string(this.rTitle.Find(infoRegion))
	description := html.UnescapeString(string(this.rRemoveHTML.ReplaceAll(
		bytes.Replace(this.rDescription.Find(infoRegion), []byte("<BR>"), []byte("\n"), -1),
		[]byte{},
	)))
	cType := InvalidComic
	switch string(this.rType.Find(infoRegion)) {
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
	for _, altTitle := range bytes.Split(this.rAltTitles.Find(infoRegion), []byte("<br />")) {
		altTitles[html.UnescapeString(string(altTitle))] = struct{}{}
	}
	statusString := string(this.rStatus.Find(infoRegion))
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
	switch string(this.rScanStatus.Find(infoRegion)) {
	case "Yes":
		scanStatus = ScanlationComplete
	case "No":
		scanStatus = ScanlationOngoing
	}
	ratingString := string(this.rRating.Find(infoRegion))
	rating, _ := strconv.ParseFloat(ratingString, 32)
	imageUrl := string(this.rImageURL.Find(infoRegion))
	image, _, _ := image.Decode(bytes.NewReader(this.fetcher().DownloadData(imageUrl)))
	genres := make(map[ComicGenreId]struct{})
	for _, genre := range qutils.Vals(ComicGenres.AssignIdsBytes(bytes.Split(this.rRemoveHTML.ReplaceAll(this.rGenres.Find(infoRegion), []byte{}), []byte("&nbsp; "))))[0].([]ComicGenreId) {
		genres[genre] = struct{}{}
	}
	categoriesAjax := this.fetcher().DownloadData("https://www.mangaupdates.com/ajax/show_categories.php?type=1&s=" + this.rComicID.FindString(url))
	categories := make(map[ComicTagId]struct{})
	for _, tag := range qutils.Vals(ComicTags.AssignIdsBytes(this.rCategories.FindAll(categoriesAjax, -1)))[0].([]ComicTagId) {
		categories[tag] = struct{}{}
	}
	authors, _ := Authors.AssignIdsBytes(this.rExtract.FindAll(this.rAuthorsLine.Find(infoRegion), -1))
	artists, _ := Artists.AssignIdsBytes(this.rExtract.FindAll(this.rArtistsLine.Find(infoRegion), -1))
	_, mature := genres[MATURE_GENRE()]

	return &ComicInfo{
		Title:            title,
		AltTitles:        altTitles,
		Authors:          authors,
		Artists:          artists,
		Genres:           genres,
		Categories:       categories,
		Type:             cType,
		Status:           status,
		ScanlationStatus: scanStatus,
		Description:      description,
		Rating:           float32(rating),
		Mature:           mature,
		Thumbnail:        image,
	}
}

func (this *BUpdates) FetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter) {
	this.initialize()
	source := comic.GetSource(this.name)
	linkPrefix := "https://www.mangaupdates.com/releases.html?stype=series&perpage=100&search=" + this.rComicID.FindString(source.URL) + "&page="
	regionsSlice := make([][]byte, 0, 20)
	regionsSlice = append(regionsSlice, this.rChaptersRegion.Find(this.fetcher().DownloadData(linkPrefix+strconv.FormatInt(1, 10))))
	pageCountString := string(this.rChpListPageCount.Find(regionsSlice[0]))
	pageCount, _ := strconv.ParseUint(pageCountString, 10, 32)
	for i := 2; i <= int(pageCount); i++ {
		regionsSlice = append(regionsSlice, this.rChaptersRegion.Find(this.fetcher().DownloadData(linkPrefix+strconv.FormatInt(int64(i), 10))))
	}
	identities = make([]ChapterIdentity, 0, pageCount*100)
	chapters = make([]Chapter, 0, pageCount*100)
	missingVolumes := false
	for i := len(regionsSlice) - 1; i >= 0; i-- { //cannot use range, because we're iterating in reverse :(
		chaptersInfos := this.rChaptersInfoList.FindAll(regionsSlice[i], -1)
		prevIdentity := ChapterIdentity{}
		for j := len(chaptersInfos) - 1; j >= 0; j-- { //I really wish they added revrange keyword
			infoPieces := this.rChpInfoPieceList.FindAll(chaptersInfos[j], -1)
			// [0] is date
			// [1] is comic title (wut)
			// [2] is volume number
			// [3] is chapter number/s
			// [4-?] is scanlators
			volumeString := string(infoPieces[2])
			missingVolumes = missingVolumes || volumeString == ""
			numberString := string(infoPieces[3])
			for i, scanlator := range infoPieces[4:] {
				infoPieces[i] = []byte(html.UnescapeString(string(scanlator)))
			}
			scanlators, _ := Scanlators.AssignIdsBytes(infoPieces[4:])
			newIdentities, _ := this.parseIdentities(volumeString, numberString, prevIdentity) //TODO: parsing error logging
			prevIdentity = newIdentities[len(newIdentities)-1]
			for _, identity := range newIdentities {
				title := "[Chapter #" + strconv.FormatInt(int64(identity.MajorNum), 10)
				if identity.MinorNum != 0 {
					title += "." + strconv.FormatInt(int64(identity.MinorNum), 10)
				}
				title += "]"
				chapter := Chapter{AlreadyRead: source.MarkAsRead}
				chapter.AddScanlation(ChapterScanlation{title, ENGLISH_LANG(), JoinScanlators(scanlators), this.name, "", make([]string, 0, 20)})

				identities = append(identities, identity)
				chapters = append(chapters, chapter)
			}
		}
	}
	if missingVolumes {
		correctiveSlice := CorrectiveSlice{identities, chapters}
		sort.Sort(correctiveSlice)
		prevVol := byte(1)
		for i := range correctiveSlice.identities {
			if correctiveSlice.identities[i].Volume == 0 {
				correctiveSlice.identities[i].Volume = prevVol
			}
			prevVol = correctiveSlice.identities[i].Volume
		}
	}
	return
}

type CorrectiveSlice struct {
	identities []ChapterIdentity
	chapters   []Chapter
}

func (this CorrectiveSlice) Len() int {
	if ilen := len(this.identities); ilen == len(this.chapters) {
		return ilen
	} else {
		return 0
	}
}

func (this CorrectiveSlice) Less(i, j int) bool {
	ident1 := this.identities[i]
	ident2 := this.identities[j]
	if ident1.Volume != 0 && ident2.Volume != 0 {
		return ident1.Less(ident2)
	} else {
		return ident1.MajorNum < ident2.MajorNum ||
			(ident1.MajorNum == ident2.MajorNum && ident1.MinorNum < ident2.MinorNum)
	}
}

func (this CorrectiveSlice) Swap(i, j int) {
	this.identities[i], this.identities[j] = this.identities[j], this.identities[i]
	this.chapters[i], this.chapters[j] = this.chapters[j], this.chapters[i]
}

func (this *BUpdates) FetchChapterPageLinks(url string) []string {
	_ = url           //unused
	return []string{} //plugin doesn't provide data, return empty list
}

func (this *BUpdates) parseIdentities(volumeStr, numberStr string, previous ChapterIdentity) (identities []ChapterIdentity, perr error) {
	errStr := strconv.Quote(volumeStr) + " + " + strconv.Quote(numberStr)
	qualityModifier := MQ_Modifier
	identity := ChapterIdentity{Version: qualityModifier + 1}
	volumeParsing := this.rIdentityParse.FindStringSubmatch(volumeStr)
	/*
		[0] is whole match
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

	numberParsing := this.rIdentityParse.FindStringSubmatch(numberStr)
	/*
		[0] is whole match
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
		identity.MinorNum += 5 //so we treat it as a special chapter
		return []ChapterIdentity{identity}, nil
	} else { //numberStr is empty, which means whole volume got scanlated, but we have no way to tell how many chapters is that
		return []ChapterIdentity{}, &CIError{errStr, nil, "Whole volume scanlated, unknown number of chapters"}
	}
}
