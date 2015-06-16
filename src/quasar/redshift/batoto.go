package redshift

import (
	"bytes"
	"html"
	"net/url"
	"path"
	"quasar/qregexp"
	"quasar/qutils"
	"quasar/qutils/qerr"
	. "quasar/redshift/idsdict"
	"quasar/redshift/qdb"
	"reflect"
	"strconv"
	"strings"
)

var (
	batoto_rURLValidator = qregexp.MustCompile(`^http://bato.to/comic/_/comics/[\w-]+r\d+$`)

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
	batoto_rImageURL        = qregexp.MustCompile(`http://img\.batoto[^"]+`)
	batoto_rExtract         = qregexp.MustCompile(`(?<=> ?)[^<]+(?=<)`)
	batoto_rExtractStrict   = qregexp.MustCompile(`(?<=">)[^<]+(?=<)`)

	batoto_rChaptersRegion   = qregexp.MustCompile(`(?s)class="ipb_table chapters_list".*</tbody>`)
	batoto_rChapterURL       = qregexp.MustCompile(`(?<=<a href=")http://bato.to/read/_/[^"]+(?="><img src=)`)
	batoto_rIdentityAndTitle = qregexp.MustCompile(`/> ([^ ]* *[^ ]+)(?: Read Online|: ([^<]+))`)
	batoto_rScanlator        = qregexp.MustCompile(`(?<=bato.to/group/_/[^"]+">)[^<]+`)
	batoto_rLang             = qregexp.MustCompile(`(?<=<div title=")[^"]+`)

	batoto_rIdentityParse = qregexp.MustCompile(`(?:Vol\.(\d+) +)?Ch\.(\d+)(?:\.(\d))?(?:v(\d))?`)

	batoto_rPageCount  = qregexp.MustCompile(`(?<=id="page_select".+\n +<.*page )\d+(?=</option>\n)`)
	batoto_rImageLink1 = qregexp.MustCompile(`(?<=id="comic_page".*src=")[^"]+`)
	batoto_rImageLink2 = qregexp.MustCompile(`(?<=img\.src = ")[^"]+`)

	batoto_rResultsRegions = qregexp.MustCompile(`http://bato.to/comic/_/.+</a>`)
	batoto_rComicURL       = qregexp.MustCompile(`http://bato.to/comic/_/[^\"]+`)
	batoto_rComicTitles    = qregexp.MustCompile(`/> ([^(<]+)(?: \(([^)]+)\))?</a>`)
)

type batoto struct {
	name      FetcherPluginName
	settings  PerPluginSettings
	m_fetcher *fetcher
}

func NewBatoto() *batoto { //TODO: logic saved as interpreted files
	ret := &batoto{}
	ret.name = FetcherPluginName(reflect.TypeOf(*ret).Name())
	return ret
}

func (this *batoto) fetcher() *fetcher {
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	return this.m_fetcher
}

func (this *batoto) setFetcher(parent *fetcher) {
	this.m_fetcher = parent
}

func (this *batoto) PluginName() FetcherPluginName {
	return this.name
}

func (this *batoto) Languages() []string {
	return []string{
		"English", "Spanish", "French", "German", "Portuguese", "Turkish", "Indonesian", "Greek", "Filipino", "Italian",
		"Polish", "Thai", "Malayan", "Hungarian", "Romanian", "Arabic", "Hebrew", "Russian", "Vietnamese", "Dutch",
	}
}

func (this *batoto) Capabilities() FetcherPluginCapabilities {
	return FetcherPluginCapabilities{
		ProvidesInfo: true,
		ProvidesData: true,
	}
}

func (this *batoto) Settings() PerPluginSettings {
	return this.settings
}

func (this *batoto) SetSettings(new PerPluginSettings) {
	if overrideMaxConns := new.OverrideDefaults[4]; overrideMaxConns {
		this.fetcher().connLimits[this.name] = new.MaxConnectionsToHost
	} else {
		this.fetcher().connLimits[this.name] = 0
	}
	this.settings = new
}

func (this *batoto) IsURLValid(url string) bool {
	return batoto_rURLValidator.MatchString(url)
}

func (this *batoto) findComicURL(title string) string {
	links, titles := this.findComicURLList(title)
	for i, ctitle := range titles {
		if strings.EqualFold(title, ctitle) {
			return links[i]
		}
	}
	return ""
}

func (this *batoto) findComicURLList(title string) (links []string, titles []string) {
	contents, err := this.fetcher().DownloadData(this.name, "http://bato.to/search?name_cond=c&name="+url.QueryEscape(title), false)
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
		links = append(links, url)
		titles = append(titles, html.UnescapeString(string(ctitles[1])))
		if str := ctitles[2]; len(str) > 0 {
			links = append(links, url)
			titles = append(titles, html.UnescapeString(string(str)))
		}
	}
	return
}

func (this *batoto) fetchComicInfo(comic *Comic) *ComicInfo {
	contents, err := this.fetcher().DownloadData(this.name, comic.GetSource(this.name).URL, true)
	if err != nil {
		panic(err)
	}
	infoRegion := batoto_rInfoRegion.Find(contents)
	title := string(batoto_rTitle.Find(infoRegion))
	altTitles := make(map[string]struct{})
	for _, altTitle := range batoto_rExtract.FindAll(batoto_rAltTitlesLine.Find(infoRegion), -1) {
		altTitles[string(altTitle)] = struct{}{}
	}
	authors, _ := Authors.AssignIdsBytes(batoto_rExtractStrict.FindAll(batoto_rAuthorsLine.Find(infoRegion), -1))
	artists, _ := Artists.AssignIdsBytes(batoto_rExtractStrict.FindAll(batoto_rArtistsLine.Find(infoRegion), -1))
	genres := make(map[ComicGenreId]struct{})
	genreNames := batoto_rExtract.FindAll(batoto_rGenresLine.Find(infoRegion), -1)
	for _, genre := range qutils.Vals(ComicGenres.AssignIdsBytes(genreNames))[0].([]ComicGenreId) {
		genres[genre] = struct{}{}
	}
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
		thumbnailFilename = path.Base(thumbnailUrl)
		thumbnail, err := this.fetcher().DownloadData(this.name, thumbnailUrl, false)
		if err != nil {
			panic(err)
		}
		qdb.SaveThumbnail(thumbnailFilename, thumbnail)
	}
	return &ComicInfo{
		Title:             title,
		AltTitles:         altTitles,
		Authors:           authors,
		Artists:           artists,
		Genres:            genres,
		Categories:        make(map[ComicTagId]struct{}), //empty
		Type:              cType,
		Status:            status,
		ScanlationStatus:  scanStatus,
		Description:       desc,
		Rating:            float32(rating),
		Mature:            mature,
		ThumbnailFilename: thumbnailFilename,
	}
}

func (this *batoto) fetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool) {
	source := comic.GetSource(this.name)
	contents, err := this.fetcher().DownloadData(this.name, source.URL, true)
	if err != nil {
		panic(err)
	}

	chaptersRegion := batoto_rChaptersRegion.Find(contents)

	chaptersList := bytes.Split(chaptersRegion, []byte("</td></tr>"))
	chaptersList = chaptersList[:len(chaptersList)-2] //Last two entries are garbage

	identities = make([]ChapterIdentity, 0, len(chaptersList))
	chapters = make([]Chapter, 0, len(chaptersList))

	for i := len(chaptersList) - 1; i >= 0; i-- { //cannot use range, because we're iterating in reverse :(
		chapterInfo := chaptersList[i]

		if bytes.HasPrefix(chapterInfo, []byte(`<tr class="chapter_row_expand`)) {
			continue // skip some bullshit they sometimes insert in the middle
		}

		lang := Langs.Id(string(batoto_rLang.Find(chapterInfo)))
		if !this.fetcher().settings.Languages[lang] {
			continue // skip disabled languages
		}

		url := string(batoto_rChapterURL.Find(chapterInfo))

		identityAndTitle := batoto_rIdentityAndTitle.FindSubmatch(chapterInfo)
		identity, _ := parseBatotoIdentity(string(identityAndTitle[1])) //TODO: log error
		missingVolumes = missingVolumes || identity.Volume == 0
		title := html.UnescapeString(string(identityAndTitle[2]))
		if title == "" {
			title = titleFromIdentity(identity)
		}

		scanlatorNames := batoto_rScanlator.FindAll(chapterInfo, -1)
		for i, scanlator := range scanlatorNames {
			scanlatorNames[i] = []byte(html.UnescapeString(string(scanlator)))
		}
		scanlators, _ := Scanlators.AssignIdsBytes(scanlatorNames)

		chapter := NewChapter(source.MarkAsRead)
		chapter.AddScanlation(ChapterScanlation{
			Title:      title,
			Language:   lang,
			Scanlators: JoinScanlators(scanlators),
			PluginName: this.name,
			URL:        url,
			PageLinks:  make([]string, 0),
		})

		identities = append(identities, identity)
		chapters = append(chapters, *chapter)
	}
	return
}

func (this *batoto) fetchChapterPageLinks(url string) []string {
	firstContents, err := this.fetcher().DownloadData(this.name, url, false)
	if err != nil {
		panic(err)
	}
	pageCount, _ := strconv.ParseUint(string(batoto_rPageCount.Find(firstContents)), 10, 8)
	contentsSlice := make([][]byte, 0, pageCount)
	contentsSlice = append(contentsSlice, firstContents)
	for i := int64(3); i <= int64(pageCount); i += 2 {
		contents, err := this.fetcher().DownloadData(this.name, url+"/"+strconv.FormatInt(i, 10), false)
		if err != nil {
			panic(err)
		}
		contentsSlice = append(contentsSlice, contents)
	}
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

func parseBatotoIdentity(str string) (identity ChapterIdentity, err error) {
	parsing := batoto_rIdentityParse.FindStringSubmatch(str)
	/*
		[0] is whole match
		[1] is volume number (optional)
		[2] is chapter major number
		[3] is chapter minor number (optional)
		[4] is chapter version (optional)
	*/
	if len(parsing) > 0 {
		if str := parsing[1]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Volume = byte(i)
		}
		/*        parsing[2]         */ {
			i, _ := strconv.ParseUint(parsing[2], 10, 16)
			identity.MajorNum = uint16(i)
		}
		if str := parsing[3]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.MinorNum = byte(i)
		}
		if str := parsing[4]; str != "" {
			i, _ := strconv.ParseUint(str, 10, 8)
			identity.Version = MQ_Modifier + byte(i)
		} else {
			identity.Version = MQ_Modifier + 1
		}
		return
	}
	return identity, qerr.NewParse("Regular expression match failed!", nil, strconv.Quote(str))
}
