package core

import (
	"bytes"
	"html"
	"net/url"
	"path"
	"reflect"
	"strconv"
	"strings"
	"sync"

	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
)

var (
	batoto_rURLValidator    = qregexp.MustCompile(`^https?://bato.to/comic/_(?:/comics)?/[\w-%]+r\d+/?$`)
	batoto_rInfoRegion      = qregexp.MustCompile(`(?s)class='rating.*class='ipsPad'`)
	batoto_rTitle           = qregexp.MustCompile(`(?<=ipsType_pagetitle'>\s+)[^\r\n]+`)
	batoto_rRating          = qregexp.MustCompile(`(?<=\()\d\.\d\d(?= - \d+votes\))`)
	batoto_rAltTitlesLine   = qregexp.MustCompile(`(?<=Alt Names:</td>\s+<td>).*(?=</td>)`)
	batoto_rAuthorsLine     = qregexp.MustCompile(`(?<=Author:</td>\s+<td>).*(?=</td>)`)
	batoto_rArtistsLine     = qregexp.MustCompile(`(?<=Artist:</td>\s+<td>).*(?=</td>)`)
	batoto_rGenresLine      = qregexp.MustCompile(`(?<=Genres:</td>\s+<td><a href[^>]+>).*(?=</a)`)
	batoto_rType            = qregexp.MustCompile(`(?<=Type:</td>\s+<td>).*(?=</)`)
	batoto_rScanStatus      = qregexp.MustCompile(`(?<=Status:</td>\s+<td>).*(?=</)`)
	batoto_rDescriptionLine = qregexp.MustCompile(`(?<=Description:</td>\s+<td>).*(?=</td>)`)
	batoto_rMature          = qregexp.MustCompile("The following content is intended for mature audiences and may contain sexual themes, gore, violence and/or strong language. Discretion is advised.")
	batoto_rImageURL        = qregexp.MustCompile(`https?://img\.bato[^"]+`)
	batoto_rExtract         = qregexp.MustCompile(`(?<=> ?)[^<]+(?=<)`)
	batoto_rExtractStrict   = qregexp.MustCompile(`(?<=">)[^<]+(?=<)`)

	batoto_rChaptersRegion = qregexp.MustCompile(`(?s)class="ipb_table chapters_list".*</tbody>`)
	//batoto_rChaptersRegion   = qregexp.MustCompile(`(?s)class="ipb_table chapters_list".*?</tbody>`)
	batoto_rChapterURL       = qregexp.MustCompile(`(?<=<a href=")https?://bato.to/read/_/[^"]+(?=" title)`)
	batoto_rIdentityAndTitle = qregexp.MustCompile(`title="(.+?)(?: ?(?:Read Online)|(?: ?: ?(.+?))) \| Sort: (\d+(?:\.\d\d?)?)\d*"><img src=`)
	batoto_rScanlator        = qregexp.MustCompile(`(?<=bato.to/group/_/[^"]+">)[^<]+`)
	batoto_rLang             = qregexp.MustCompile(`(?<=<div title=")[^"]+`)

	batoto_rStrictIdentityParse = qregexp.MustCompile(`^(?:Vol\.(\d+) +)?Ch\.(\d+)(?:(?:\.(\d))?-?([a-gA-G])?|-(\d+))(?:(?: (?:\(|\[)?)?v\.?(\d).?)?$`)
	batoto_rGuessIdentityParse  = qregexp.MustCompile(`(?:Vol\.(\d+) +)?Ch\.(?:(\d+)|[\w]+)(?:(?:\.(\d))?-?([a-gA-G])?|-(\d+))(?:(?: (?:\(|\[)?)?v\.?(\d).?)?`)
	batoto_rIsColor             = qregexp.MustCompile(`[Cc]olor`)
	batoto_rSpecialRule_1_Parse = qregexp.MustCompile(`^Ch\.v(\d+)`)
	batoto_rSpecialRule_2_Parse = qregexp.MustCompile(`^Vol\.(\d+)\.(\d+)`)

	batoto_rPageCount  = qregexp.MustCompile(`(?<=id="page_select".+\n +<.*page )\d+(?=</option>\n)`)
	batoto_rImageLink1 = qregexp.MustCompile(`(?<=id="comic_page".*src=")[^"]+`)
	batoto_rImageLink2 = qregexp.MustCompile(`(?<=img\.src = ")[^"]+`)

	batoto_rResultsRegions = qregexp.MustCompile(`https?://bato.to/comic/_/.+</a>`)
	batoto_rComicURL       = qregexp.MustCompile(`https?://bato.to/comic/_/[^\"]+`)
	batoto_rComicTitles    = qregexp.MustCompile(`/> ([^(<]+(?: \([A-Z]+ [^)]+\))?)(?i: \(Doujinshi\))?(?: \((.[^)]+)\))?</a>`)
)

type batoto struct {
	sourceSharedImpl
}

func NewBatoto() *batoto { //TODO: logic saved as interpreted files
	ret := &batoto{}
	ret.id = SourceId(reflect.TypeOf(*ret).Name())
	return ret
}

func (this *batoto) Name() string {
	return "Batoto"
}

func (this *batoto) Languages() []string {
	return []string{
		"English", "Spanish", "French", "German", "Portuguese", "Turkish", "Indonesian", "Greek", "Filipino", "Italian",
		"Polish", "Thai", "Malayan", "Hungarian", "Romanian", "Arabic", "Hebrew", "Russian", "Vietnamese", "Dutch",
	}
}

func (this *batoto) Capabilities() SourceCapabilities {
	return SourceCapabilities{
		ProvidesMetadata: true,
		ProvidesData:     true,
	}
}

func (this *batoto) IsURLValid(url string) bool {
	return batoto_rURLValidator.MatchString(url)
}

func (this *batoto) advert() advert {
	return advert{} //TODO
}

func (this *batoto) search(title, author string, genres []ComicGenreId, status comicStatus, ctype comicType, mature bool) []comicSearchResult {
	return []comicSearchResult(nil)
}

func (this *batoto) comicURL(title string) string {
	links, titles := this.findComicURLList(title)
	for i, ctitle := range titles {
		if strings.EqualFold(title, ctitle) {
			return links[i]
		}
	}
	return ""
}

func (this *batoto) findComicURLList(title string) (links []string, titles []string) {
	contents, err := this.fetcher().DownloadData(this.id, "http://bato.to/search?name_cond=c&name="+url.QueryEscape(title), false)
	if err != nil {
		panic(err)
	}
	urlAndTitlesList := batoto_rResultsRegions.FindAll(contents, -1)
	for _, urlAndTitles := range urlAndTitlesList {
		url := string(batoto_rComicURL.Find(urlAndTitles))
		ctitles := batoto_rComicTitles.FindSubmatch(urlAndTitles)
		/*
			[0] is entire match
			[1] is first title
			[2] is second title (optional)
		*/
		if ctitles == nil {
			qlog.Log(qlog.Warning, "nil urlAndTitles", string(urlAndTitles))
			continue
		}
		links = append(links, url)
		titles = append(titles, html.UnescapeString(string(ctitles[1])))
		if str := ctitles[2]; len(str) > 0 {
			links = append(links, url)
			titles = append(titles, html.UnescapeString(string(str)))
		}
	}
	return
}

func (this *batoto) comicInfo(source SourceLink) *ComicInfo {
	if source.SourceId != this.id {
		panic("Incompatible SourceLink of " + string(source.SourceId) + "::" + source.URL + " provided!")
	}

	contents, err := this.fetcher().DownloadData(this.id, source.URL, true)
	if err != nil {
		panic(err)
	}

	infoRegion := batoto_rInfoRegion.Find(contents)
	title := html.UnescapeString(string(batoto_rTitle.Find(infoRegion)))
	titles := append([]string{title}, qutils.ByteSlicesToStrings(batoto_rExtract.FindAll(batoto_rAltTitlesLine.Find(infoRegion), -1))...)

	authors, _ := this.fetcher().authors().AssignIdsBytes(batoto_rExtractStrict.FindAll(batoto_rAuthorsLine.Find(infoRegion), -1))
	artists, _ := this.fetcher().artists().AssignIdsBytes(batoto_rExtractStrict.FindAll(batoto_rArtistsLine.Find(infoRegion), -1))
	genres, _ := this.fetcher().genres().AssignIdsBytes(batoto_rExtract.FindAll(batoto_rGenresLine.Find(infoRegion), -1))

	cType := InvalidComic
	switch string(batoto_rType.Find(infoRegion)) {
	case "Manga (Japanese)":
		cType = Manga
	case "Manhwa (Korean)":
		cType = Manhwa
	case "Manhua (Chinese)":
		cType = Manhua
	case "Other":
		cType = Western
	}

	status := ComicStatusInvalid
	scanStatus := ScanlationStatusInvalid
	switch string(batoto_rScanStatus.Find(infoRegion)) {
	case "Ongoing":
		scanStatus = ScanlationOngoing
		status = ComicOngoing
	case "Complete":
		scanStatus = ScanlationComplete
		status = ComicComplete
	}

	desc := html.UnescapeString(string(bytes.Replace(batoto_rDescriptionLine.Find(infoRegion), []byte("<br />"), []byte("\n"), -1)))
	mature := batoto_rMature.Match(infoRegion)
	rating, _ := strconv.ParseFloat(string(batoto_rRating.Find(infoRegion)), 32)

	var thumbnailFilename string
	thumbnailUrl := string(batoto_rImageURL.Find(infoRegion))
	if thumbnailUrl != "" {
		if thumbnailFilename = path.Base(thumbnailUrl); !qdb.ThumbnailExists(thumbnailFilename) {
			thumbnail, err := this.fetcher().DownloadData(this.id, thumbnailUrl, false)
			if err != nil {
				panic(err)
			}
			qdb.SaveThumbnail(thumbnailFilename, thumbnail)
		}
	}

	return &ComicInfo{
		MainTitleIdx:     0,
		Titles:           titles,
		Authors:          authors,
		Artists:          artists,
		Genres:           genres,
		Categories:       []ComicTagId(nil), //empty
		Type:             cType,
		Status:           status,
		ScanlationStatus: scanStatus,
		Description:      desc,
		Rating:           uint16(rating * 200), // x/5 * 10 * 100 (e.g. 4.81/5 * 10 = 9.62 on a 10pt scale)
		Mature:           mature,
		ThumbnailIdx:     0,
		Thumbnails:       []string{thumbnailFilename},
	}
}

func (this *batoto) chapterList(source SourceLink) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool) {
	if source.SourceId != this.id {
		panic("Incompatible SourceLink of " + string(source.SourceId) + "::" + source.URL + " provided!")
	}

	contents, err := this.fetcher().DownloadData(this.id, source.URL, true)
	if err != nil {
		panic(err)
	}

	chaptersRegion := batoto_rChaptersRegion.Find(contents)

	chaptersList := bytes.Split(chaptersRegion, []byte("</td></tr>"))
	chaptersList = chaptersList[:len(chaptersList)-2] //Last two entries are garbage

	identities = make([]ChapterIdentity, 0, len(chaptersList))
	chapters = make([]Chapter, 0, len(chaptersList))

	cachedComicTitle := "" //in case of errors
	comicTitle := func() string {
		if cachedComicTitle == "" {
			infoRegion := batoto_rInfoRegion.Find(contents)
			cachedComicTitle = html.UnescapeString(string(batoto_rTitle.Find(infoRegion)))
		}
		return cachedComicTitle
	}

	for i := len(chaptersList) - 1; i >= 0; i-- { //cannot use range, because we're iterating in reverse :(
		chapterInfo := chaptersList[i]

		if bytes.HasPrefix(chapterInfo, []byte(`<tr class="chapter_row_expand`)) {
			continue // skip some bullshit they sometimes insert in the middle
		}

		langsDict := this.fetcher().langs()
		lang := langsDict.Id(string(batoto_rLang.Find(chapterInfo)))
		if !this.fetcher().settings.Languages[LangName(langsDict.NameOf(lang))] {
			continue // skip disabled languages
		}

		url := string(batoto_rChapterURL.Find(chapterInfo))

		identityAndTitle := batoto_rIdentityAndTitle.FindSubmatch(chapterInfo)
		if identityAndTitle == nil {
			qlog.Logf(qlog.Error, `Failed to extract %s chapter identity and title for comic %s`, this.Name(), comicTitle())
			qlog.Logf(qlog.Error, "\n%s\n", string(chapterInfo))
			continue
		}
		idStr, sortHint := string(identityAndTitle[1]), string(identityAndTitle[3])
		identity, strict, color, version, err := parseBatotoIdentity(idStr, sortHint)
		if err != nil {
			qlog.Logf(qlog.Error, "Parsing %s chapter identity for comic \"%s\" failed: %v", this.Name(), comicTitle(), err)
			continue
		}
		if !strict {
			qlog.Logf(qlog.Warning, "Irregular %s chapter identity \"%s | %s\" for comic \"%s\"; parsed as %v",
				this.Name(), idStr, sortHint, comicTitle(), identity,
			)
		}

		missingVolumes = missingVolumes || identity.Volume == 0
		title := html.UnescapeString(string(identityAndTitle[2]))
		if title == "" {
			title = titleFromIdentity(identity)
		}

		scanlatorNames := batoto_rScanlator.FindAll(chapterInfo, -1)
		for i, scanlator := range scanlatorNames {
			scanlatorNames[i] = []byte(html.UnescapeString(string(scanlator)))
		}
		scanlators, _ := this.fetcher().scanlators().AssignIdsBytes(scanlatorNames)

		chapter := Chapter{MarkedRead: source.MarkAsRead}
		chapter.AddScanlation(ChapterScanlation{
			SourceId:    this.id,
			Scanlators:  JoinScanlators(scanlators),
			Version:     version,
			Color:       color,
			Title:       title,
			Language:    lang,
			MetadataURL: url,
			PageLinks:   make([]string, 0),
		})

		identities = append(identities, identity)
		chapters = append(chapters, chapter)
	}
	return
}

func (this *batoto) chapterDataLinks(url string) []string { //TODO: also handle single-page-multiple-images chapters
	firstContents, err := this.fetcher().DownloadData(this.id, url, false)
	if err != nil {
		panic(err)
	}

	pageCount, _ := strconv.ParseUint(string(batoto_rPageCount.Find(firstContents)), 10, 8)
	contentsSlice := make([][]byte, (pageCount+1)/2) //We don't have to download all the pages, they also contain a link to the next image
	contentsSlice[0] = firstContents
	idx := 1
	var wg sync.WaitGroup
	for i := int64(3); i <= int64(pageCount); i += 2 {
		wg.Add(1)
		go func(sliceIdx int, pageIdx int64) {
			defer wg.Done()
			contents, err := this.fetcher().DownloadData(this.id, url+"/"+strconv.FormatInt(pageIdx, 10), false)
			if err != nil {
				panic(err)
			}
			contentsSlice[sliceIdx] = contents
		}(idx, i)
		idx++
	}
	wg.Wait()

	pageLinks := make([]string, 0, pageCount)
	for _, contents := range contentsSlice {
		pageLinks = append(pageLinks, string(batoto_rImageLink1.Find(contents)))
		imageLink2 := batoto_rImageLink2.Find(contents)
		if len(imageLink2) > 0 {
			pageLinks = append(pageLinks, string(imageLink2))
		}
	}
	return pageLinks
}

func parseBatotoIdentity(idStr, sortHint string) (identity ChapterIdentity, strict, color bool, version byte, err error) {
	normalizeNumbers := func(r rune) rune {
		if r >= '０' && r <= '９' { //should probably also handle the rest of unicode digits,
			return r - '０' + '0' //but they're very unlikely to show up in practice
		}
		return r
	}
	idStr = strings.Map(normalizeNumbers, idStr)
	color = batoto_rIsColor.MatchString(idStr)

	strict = true
	parsing := batoto_rStrictIdentityParse.FindStringSubmatch(idStr)
	if parsing == nil {
		if workaround := batoto_rSpecialRule_1_Parse.FindStringSubmatch(idStr); workaround != nil {
			parsing = []string{workaround[0], workaround[1], "0", "", "", "", ""}
		} else if workaround = batoto_rSpecialRule_2_Parse.FindStringSubmatch(idStr); workaround != nil {
			parsing = []string{workaround[0], workaround[1], "", workaround[2], "", "", ""}
		} else {
			parsing = batoto_rGuessIdentityParse.FindStringSubmatch(idStr)
		}
		strict = false
	}
	if parsing != nil {
		/*
			[0] is whole match
			[1] is volume number (optional)
			[2] is chapter major number
			[3] is chapter minor number (optional)
			[4] is chapter part letter (optional)
			[5] is chapter range (optional, unhandled) or chapter part number (malformed, must be lesser than or equal [2])
			[6] is chapter version (optional)
		*/
		if str := parsing[1]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Volume = byte(i)
		}
		if str := parsing[2]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 16)
			identity.MajorNum = uint16(i)
		}
		if str := parsing[3]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.MinorNum = byte(i)
		}
		if str := parsing[4]; str != "" {
			letter := str[0]
			if letter < 'a' {
				letter += ('a' - 'A')
			}
			identity.Letter = byte(letter - 'a' + 1)
		}
		if str := parsing[5]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 16)
			if uint16(i) <= identity.MajorNum {
				identity.Letter = byte(i)
			} else {
				//no way to handle chapter ranges, ignore
			}
		}
		if str := parsing[6]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			version = byte(i)
		} else {
			version = 1
		}

		if identity.Volume != 0 && strict {
			return identity, strict, color, version, nil
		}
	}

	hint := strings.Split(sortHint, ".")
	if len(hint) < 2 {
		hint = append(hint, "0")
	}
	majorHint, minorHint := hint[0], hint[1]
	if len(minorHint) > 1 && minorHint[1] == '9' {
		switch minorHint[0] {
		case '0':
			minorHint = "1"
		case '1':
			minorHint = "2"
		case '2':
			minorHint = "3"
		case '3':
			minorHint = "4"
		case '4':
			minorHint = "5"
		case '5':
			minorHint = "6"
		case '6':
			minorHint = "7"
		case '7':
			minorHint = "8"
		case '8':
			minorHint = "9"
		}
	}

	if len(majorHint) > 3 {
		i, err := strconv.ParseUint(majorHint[:len(majorHint)-3], 10, 8)
		if err != nil {
			return ChapterIdentity{}, false, false, 0, qerr.NewParse("Malformed sort hint (volume part)", err, strconv.Quote(sortHint))
		}
		identity.Volume = byte(i)
		majorHint = majorHint[len(majorHint)-3:]
	}

	if strict {
		return identity, strict, color, version, nil
	}

	if identity.MajorNum == 0 {
		i, err := strconv.ParseUint(majorHint, 10, 16)
		if err != nil {
			return ChapterIdentity{}, false, false, 0, qerr.NewParse("Malformed sort hint (chapter-major part)", err, strconv.Quote(sortHint))
		}
		identity.MajorNum = uint16(i)
	}

	if identity.MinorNum == 0 {
		i, err := strconv.ParseUint(minorHint[:1], 10, 8)
		if err != nil {
			return ChapterIdentity{}, false, false, 0, qerr.NewParse("Malformed sort hint (chapter-minor part)", err, strconv.Quote(sortHint))
		}
		if i == 0 && !strings.ContainsRune(idStr, '+') &&
			(strings.Contains(idStr, "Extra") || strings.Contains(idStr, "Special") || strings.Contains(idStr, "Omake")) {
			i = 5
		}
		identity.MinorNum = byte(i)
	}

	if version == 0 {
		version = 1
	}

	return identity, strict, color, version, nil
}
