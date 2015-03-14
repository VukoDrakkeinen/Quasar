package redshift

import (
	"bytes"
	"fmt"
	"html"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"quasar/qregexp"
	"quasar/qutils"
	. "quasar/redshift/idbase"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type Batoto struct {
	initialized bool
	name        FetcherPluginName
	m_fetcher   *Fetcher

	rURLValidator *qregexp.QRegexp

	rInfoRegion      *qregexp.QRegexp
	rTitle           *qregexp.QRegexp
	rRating          *qregexp.QRegexp
	rAltTitlesLine   *qregexp.QRegexp
	rAuthorsLine     *qregexp.QRegexp
	rArtistsLine     *qregexp.QRegexp
	rGenresLine      *qregexp.QRegexp
	rType            *qregexp.QRegexp
	rScanStatus      *qregexp.QRegexp
	rDescriptionLine *qregexp.QRegexp
	rMature          *qregexp.QRegexp
	rImageURL        *qregexp.QRegexp
	rExtract         *qregexp.QRegexp
	rExtractStrict   *qregexp.QRegexp

	rChaptersRegion   *qregexp.QRegexp
	rLang             *qregexp.QRegexp
	rChapterURL       *qregexp.QRegexp
	rIdentityAndTitle *qregexp.QRegexp
	rScanlator        *qregexp.QRegexp

	rIdentityParse *qregexp.QRegexp

	rPageCount  *qregexp.QRegexp
	rImageLink1 *qregexp.QRegexp
	rImageLink2 *qregexp.QRegexp

	rResultsRegions *qregexp.QRegexp
	rComicURL       *qregexp.QRegexp
	rComicTitles    *qregexp.QRegexp
}

func NewBatoto() *Batoto {
	return new(Batoto).initialize()
}

func (this *Batoto) initialize() *Batoto { //TODO: handle errors
	if !this.initialized { //TODO: logic saved as interpreted files

		this.name = FetcherPluginName(reflect.TypeOf(*this).Name())

		this.rURLValidator = qregexp.MustCompile(`^http://bato.to/comic/_/comics/[\w-]+r\d+$`)

		this.rInfoRegion = qregexp.MustCompile(`(?s)class='rating.*class='ipsPad'`)
		this.rTitle = qregexp.MustCompile(`(?<=ipsType_pagetitle'>\s+)[^\r\n]+`)
		this.rRating = qregexp.MustCompile(`(?<=\()\d\.\d\d(?= - \d+votes\))`)
		this.rAltTitlesLine = qregexp.MustCompile(`(?<=Alt Names:</td>\s+<td>).*(?=</td>)`)
		this.rAuthorsLine = qregexp.MustCompile(`(?<=Author:</td>\s+<td>).*(?=</td>)`)
		this.rArtistsLine = qregexp.MustCompile(`(?<=Artist:</td>\s+<td>).*(?=</td>)`)
		this.rGenresLine = qregexp.MustCompile(`(?<=Genres:</td>\s+<td><a href[^>]+>).*(?=</a)`)
		this.rType = qregexp.MustCompile(`(?<=Type:</td>\s+<td>).*(?=</)`)
		this.rScanStatus = qregexp.MustCompile(`(?<=Status:</td>\s+<td>).*(?=</)`)
		this.rDescriptionLine = qregexp.MustCompile(`(?<=Description:</td>\s+<td>).*(?=</td>)`)
		this.rMature = qregexp.MustCompile("The following content is intended for mature audiences and may contain sexual themes, gore, violence and/or strong language. Discretion is advised.")
		this.rImageURL = qregexp.MustCompile(`http://img\.batoto[^"]+`)
		this.rExtract = qregexp.MustCompile(`(?<=> ?)[^<]+(?=<)`)
		this.rExtractStrict = qregexp.MustCompile(`(?<=">)[^<]+(?=<)`)

		this.rChaptersRegion = qregexp.MustCompile(`(?s)class="ipb_table chapters_list".*</tbody>`)
		this.rChapterURL = qregexp.MustCompile(`(?<=<a href=")http://bato.to/read/_/[^"]+(?="><img src=)`)
		this.rIdentityAndTitle = qregexp.MustCompile(`/> ([^ ]* *[^ ]+)(?: Read Online|: ([^<]+))`)
		this.rScanlator = qregexp.MustCompile(`(?<=bato.to/group/_/[^"]+">)[^<]+`)
		this.rLang = qregexp.MustCompile(`(?<=<div title=")[^"]+`)

		this.rIdentityParse = qregexp.MustCompile(`(?:Vol\.(\d+) +)?Ch\.(\d+)(?:\.(\d))?(?:v(\d))?`)

		this.rPageCount = qregexp.MustCompile(`(?<=id="page_select".+\n +<.*page )\d+(?=</option>\n)`)
		this.rImageLink1 = qregexp.MustCompile(`(?<=id="comic_page".*src=")[^"]+`)
		this.rImageLink2 = qregexp.MustCompile(`(?<=img\.src = ")[^"]+`)

		this.rResultsRegions = qregexp.MustCompile(`http://bato.to/comic/_/.+</a>`)
		this.rComicURL = qregexp.MustCompile(`http://bato.to/comic/_/[^\"]+`)
		this.rComicTitles = qregexp.MustCompile(`/> ([^(<]+)(?: \(([^)]+)\))?</a>`)

		this.initialized = true
		fmt.Println("Plugin", this.name, "initialized!")
	}
	return this
}

func (this *Batoto) fetcher() *Fetcher {
	if this.m_fetcher == nil {
		panic("Fetcher is nil!")
	}
	return this.m_fetcher
}

func (this *Batoto) SetFetcher(parent *Fetcher) {
	this.initialize()
	this.m_fetcher = parent
}

func (this *Batoto) PluginName() FetcherPluginName {
	this.initialize()
	return this.name
}

func (this *Batoto) Languages() []string {
	return []string{
		"English", "Spanish", "French", "German", "Portuguese", "Turkish", "Indonesian", "Greek", "Filipino", "Italian",
		"Polish", "Thai", "Malayan", "Hungarian", "Romanian", "Arabic", "Hebrew", "Russian", "Vietnamese", "Dutch",
	}
}

func (this *Batoto) Capabilities() FetcherPluginCapabilities {
	this.initialize()
	return FetcherPluginCapabilities{
		ProvidesInfo: true,
		ProvidesData: true,
	}
}

func (this *Batoto) IsURLValid(url string) bool {
	return this.rURLValidator.MatchString(url)
}

func (this *Batoto) FindComicURL(title string) string {
	this.initialize()
	links, titles := this.FindComicURLList(title)
	for i, ctitle := range titles {
		if strings.EqualFold(title, ctitle) {
			return links[i]
		}
	}
	return ""
}

func (this *Batoto) FindComicURLList(title string) (links []string, titles []string) {
	this.initialize()
	contents := this.fetcher().DownloadData("http://bato.to/search?name_cond=c&name=" + url.QueryEscape(title))
	urlAndTitlesList := this.rResultsRegions.FindAll(contents, -1)
	for _, urlAndTitles := range urlAndTitlesList {
		url := string(this.rComicURL.Find(urlAndTitles))
		ctitles := this.rComicTitles.FindSubmatch(urlAndTitles)
		/*
			[0] is whole match
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

func (this *Batoto) FetchComicInfo(comic *Comic) *ComicInfo {
	this.initialize()
	contents := this.fetcher().DownloadData(comic.GetSource(this.name).URL)
	infoRegion := this.rInfoRegion.Find(contents)
	title := string(this.rTitle.Find(infoRegion))
	altTitles := make(map[string]struct{})
	for _, altTitle := range this.rExtract.FindAll(this.rAltTitlesLine.Find(infoRegion), -1) {
		altTitles[string(altTitle)] = struct{}{}
	}
	authors, _ := Authors.AssignIdsBytes(this.rExtractStrict.FindAll(this.rAuthorsLine.Find(infoRegion), -1))
	artists, _ := Artists.AssignIdsBytes(this.rExtractStrict.FindAll(this.rArtistsLine.Find(infoRegion), -1))
	genres := make(map[ComicGenreId]struct{})
	genreNames := this.rExtract.FindAll(this.rGenresLine.Find(infoRegion), -1)
	for _, genre := range qutils.Vals(ComicGenres.AssignIdsBytes(genreNames))[0].([]ComicGenreId) {
		genres[genre] = struct{}{}
	}
	cType := InvalidComic
	switch string(this.rType.Find(infoRegion)) {
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
	switch string(this.rScanStatus.Find(infoRegion)) {
	case "Ongoing":
		scanStatus = ScanlationOngoing
		status = ComicOngoing
	case "Complete":
		scanStatus = ScanlationComplete
		status = ComicComplete
	}
	desc := html.UnescapeString(string(bytes.Replace(this.rDescriptionLine.Find(infoRegion), []byte("<br />"), []byte("\n"), -1)))
	mature := this.rMature.Match(infoRegion)
	rating, _ := strconv.ParseFloat(string(this.rRating.Find(infoRegion)), 32)
	image, _, _ := image.Decode(bytes.NewReader(this.fetcher().DownloadData(string(this.rImageURL.Find(infoRegion)))))
	return &ComicInfo{
		Title:            title,
		AltTitles:        altTitles,
		Authors:          authors,
		Artists:          artists,
		Genres:           genres,
		Categories:       make(map[ComicTagId]struct{}), //empty
		Type:             cType,
		Status:           status,
		ScanlationStatus: scanStatus,
		Description:      desc,
		Rating:           float32(rating),
		Mature:           mature,
		Thumbnail:        image,
	}
}

func (this *Batoto) FetchChapterList(comic *Comic) (identities []ChapterIdentity, chapters []Chapter) {
	this.initialize()
	source := comic.GetSource(this.name)
	link := source.URL
	contents := this.fetcher().DownloadData(link) //TODO: cache

	chaptersRegion := this.rChaptersRegion.Find(contents)

	chaptersList := bytes.Split(chaptersRegion, []byte("</td></tr>"))
	chaptersList = chaptersList[:len(chaptersList)-2] //Last two entries are garbage

	identities = make([]ChapterIdentity, 0, len(chaptersList))
	chapters = make([]Chapter, 0, len(chaptersList))

	missingVolumes := false
	//previousIdentity := ChapterIdentity{Volume: 1} //start with volume #1
	for i := len(chaptersList) - 1; i >= 0; i-- { //cannot use range, because we're iterating in reverse :(
		chapterInfo := chaptersList[i]

		if bytes.HasPrefix(chapterInfo, []byte(`<tr class="chapter_row_expand`)) {
			continue // skip some bullshit they sometimes insert in the middle
		}

		lang := LangDict.Id(string(this.rLang.Find(chapterInfo)))
		if !this.fetcher().settings.Languages[lang] {
			continue // skip disabled languages
		}

		url := string(this.rChapterURL.Find(chapterInfo))

		identityAndTitle := this.rIdentityAndTitle.FindSubmatch(chapterInfo)
		identity, _ := this.parseIdentity(string(identityAndTitle[1])) //TODO: log error
		//identity, err := this.parseIdentity(string(identityAndTitle[1]), previousIdentity) //TODO: log error
		//if err == nil {
		//	previousIdentity = identity
		//}
		missingVolumes = missingVolumes || identity.Volume == 0
		title := html.UnescapeString(string(identityAndTitle[2]))
		if title == "" { //TODO: shared plugin logic?
			title = "[Chapter #" + strconv.FormatInt(int64(identity.MajorNum), 10)
			if identity.MinorNum != 0 {
				title += "." + strconv.FormatInt(int64(identity.MinorNum), 10)
			}
			title += "]"
		}

		scanlatorNames := this.rScanlator.FindAll(chapterInfo, -1)
		for i, scanlator := range scanlatorNames {
			scanlatorNames[i] = []byte(html.UnescapeString(string(scanlator)))
		}
		scanlators, _ := Scanlators.AssignIdsBytes(scanlatorNames)

		chapter := Chapter{AlreadyRead: source.MarkAsRead}
		chapter.AddScanlation(ChapterScanlation{title, lang, JoinScanlators(scanlators), this.name, url, make([]string, 0, 20), InDBStatusHolder{}})

		identities = append(identities, identity)
		chapters = append(chapters, chapter)
	}
	if missingVolumes { //TODO: shared plugin logic
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

func (this *Batoto) FetchChapterPageLinks(url string) []string {
	this.initialize()
	firstContents := this.fetcher().DownloadData(url)
	pageCount, _ := strconv.ParseUint(string(this.rPageCount.Find(firstContents)), 10, 8)
	contentsSlice := make([][]byte, 0, pageCount)
	contentsSlice = append(contentsSlice, firstContents)
	for i := int64(3); i <= int64(pageCount); i += 2 {
		contentsSlice = append(contentsSlice, this.fetcher().DownloadData(url+"/"+strconv.FormatInt(i, 10)))
		fmt.Println("PageSourceLink:", url+"/"+strconv.FormatInt(i, 10))
	}
	pageLinks := make([]string, 0, pageCount)
	for _, contents := range contentsSlice {
		pageLinks = append(pageLinks, string(this.rImageLink1.Find(contents)))
		imageLink2 := this.rImageLink2.Find(contents)
		if len(imageLink2) > 0 {
			pageLinks = append(pageLinks, string(imageLink2))
		}
	}
	return pageLinks
}

type CIError struct {
	Input string
	Err   error //caused by
	msg   string
}

func (this *CIError) Error() string {
	return this.msg + " (caused by: " + this.Err.Error() + ")"
}

func (this *Batoto) parseIdentity(str string /*, previous ChapterIdentity*/) (identity ChapterIdentity, perr error) {
	parsing := this.rIdentityParse.FindStringSubmatch(str)
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
		} //else {
		//	identity.Volume = previous.Volume
		//}
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
		return identity, nil
	} else {
		return identity, &CIError{strconv.Quote(str), nil, "Regular expression matching failed!"}
	}
}
