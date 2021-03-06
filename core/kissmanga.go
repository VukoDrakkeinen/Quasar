package core

import (
	"bytes"
	. "github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qregexp"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"html"
	neturl "net/url"
	"path"
	"reflect"
	"strconv"
)

var (
	kissmanga_rURLValidator = qregexp.MustCompile(`^https?://kissmanga.com/Manga/[\w-]+$`)
	kissmanga_rInfoRegion   = qregexp.MustCompile(`<a Class="bigChar" href="/Manga/(?s:.+?)<div id="divAds"`)
	kissmanga_rMainTitle    = qregexp.MustCompile(`(?<=^[^>]+?>)[^<]+(?=</a>)`)
	kissmanga_rTitlesLine   = qregexp.MustCompile(`<span class="info">Other name:.+`)
	kissmanga_rGenresLine   = qregexp.MustCompile(`<span class="info">Genres:.+`)
	kissmanga_rAuthorsLine  = qregexp.MustCompile(`<span class="info">Author:.+`)
	kissmanga_rStatus       = qregexp.MustCompile(`<span class="info">Status:.+?&nbsp;(.+)`)
	kissmanga_rDescription  = qregexp.MustCompile(`(?<=<span class="info">Summary:.+\n).+`)
	kissmanga_rImageURL     = qregexp.MustCompile(`(?<=<link rel="image_src" href=")[^"]+`)
	kissmanga_rExtract      = qregexp.MustCompile(`(?<=">)[^<]+(?=</a>)`)

	kissmanga_rChapter = qregexp.MustCompile(`<a href="(/Manga/[^"]+)".+\n.+?(?:Ch\. ?)?(\d+)(?:\.(\d+))?(?: ?:| -)? ?(?:Read Online|([^<]*))</a>`)

	kissmanga_rPageLinks    = qregexp.MustCompile(`(?<=lstImages.push\(")[^"]+(?="\);)`)
	kissmanga_rProxiedImgur = qregexp.MustCompile(`(?<=^https://images2-focus-opensocial.googleusercontent.com/gadgets/proxy\?container=focus&gadget=a&no_expand=1&resize_h=0&rewriteMime=image%2F\*&url=)[^&]+`)
)

type kissmanga struct {
	sourceSharedImpl
}

func NewKissManga() *kissmanga { //TODO: logic saved as interpreted files
	ret := &kissmanga{}
	ret.id = SourceId(reflect.TypeOf(*ret).Name())
	return ret
}

func (this *kissmanga) Name() string {
	return "KissManga"
}

func (this *kissmanga) Languages() []string {
	return []string{"English"}
}

func (this *kissmanga) Capabilities() SourceCapabilities {
	return SourceCapabilities{
		ProvidesMetadata: true,
		ProvidesData:     true,
	}
}

func (this *kissmanga) IsURLValid(url string) bool {
	return kissmanga_rURLValidator.MatchString(url)
}

func (this *kissmanga) advert() advert {
	return advert{} //TODO
}

func (this *kissmanga) search(title, author string, genres []ComicGenreId, status comicStatus, ctype comicType, mature bool) []comicSearchResult {
	return []comicSearchResult(nil)
}

func (this *kissmanga) comicURL(title string) string {
	return ""
}

func (this *kissmanga) comicInfo(source SourceLink) *ComicInfo {
	if source.SourceId != this.id {
		panic("Incompatible SourceLink of " + string(source.SourceId) + "::" + source.URL + " provided!")
	}

	contents, err := this.fetcher().DownloadData(this.id, source.URL, true)
	if err != nil {
		panic(err)
	}

	infoRegion := kissmanga_rInfoRegion.Find(contents)
	mainTitle := html.UnescapeString(string(kissmanga_rMainTitle.Find(infoRegion)))
	titles := append([]string{mainTitle}, qutils.ByteSlicesToStrings(kissmanga_rExtract.FindAll(kissmanga_rTitlesLine.Find(infoRegion), -1))...)

	aAA := kissmanga_rExtract.FindAll(kissmanga_rAuthorsLine.Find(infoRegion), -1)
	var authors []AuthorId
	var artists []ArtistId
	if alen := len(aAA); alen > 1 {
		artists, _ = this.fetcher().artists().AssignIdsBytes([][]byte{aAA[1]})
	} else if alen > 0 {
		authors, _ = this.fetcher().authors().AssignIdsBytes([][]byte{aAA[0]})
	}

	genreNames := kissmanga_rExtract.FindAll(kissmanga_rGenresLine.Find(infoRegion), -1)
	genres, _ := this.fetcher().genres().AssignIdsBytes(genreNames)

	cType := InvalidComic
	for _, genre := range genreNames { //TODO
		switch string(genre) {
		case "Manga":
			cType = Manga
		case "Manhwa":
			cType = Manhwa
		case "Manhua":
			cType = Manhua
		default:
			continue
		}
		break
	}

	status := ComicStatusInvalid
	scanStatus := ScanlationStatusInvalid
	switch string(kissmanga_rStatus.Find(infoRegion)) {
	case "Ongoing":
		scanStatus = ScanlationOngoing
		status = ComicOngoing
	case "Complete":
		scanStatus = ScanlationComplete
		status = ComicComplete
	}

	desc := html.UnescapeString(
		string(shared_rRemoveHTML.ReplaceAllLiteral(
			bytes.Replace(kissmanga_rDescription.Find(infoRegion), []byte("<br/>"), []byte("\n"), -1),
			[]byte(nil),
		)),
	)
	mature := qutils.Contains(genres, MATURE_GENRE_ID())

	var thumbnailFilename string
	thumbnailUrl := string(kissmanga_rImageURL.Find(contents))
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
		Rating:           0,
		Mature:           mature,
		ThumbnailIdx:     0,
		Thumbnails:       []string{thumbnailFilename},
	}
}

func (this *kissmanga) chapterList(source SourceLink) (identities []ChapterIdentity, chapters []Chapter, missingVolumes bool) {
	if source.SourceId != this.id {
		panic("Incompatible SourceLink of " + string(source.SourceId) + "::" + source.URL + " provided!")
	}

	contents, err := this.fetcher().DownloadData(this.id, source.URL, true)
	if err != nil {
		panic(err)
	}

	/*
		[0] is entire match
		[1] is URL
		[2] is major number
		[3] is minor number (optional)
		[4] is title
	*/
	chaptersList := kissmanga_rChapter.FindAllSubmatch(contents, -1)
	identities = make([]ChapterIdentity, 0, len(chaptersList))
	chapters = make([]Chapter, 0, len(chaptersList))

	for i := len(chaptersList) - 1; i >= 0; i-- {
		chapterInfo := chaptersList[i]
		url := "http://kissmanga.com" + string(chapterInfo[1])
		majorNum, _ := strconv.ParseUint(string(chapterInfo[2]), 10, 16)
		minorNum, _ := strconv.ParseUint(string(chapterInfo[3]), 10, 8)
		identity := ChapterIdentity{MajorNum: uint16(majorNum), MinorNum: byte(minorNum)}
		title := html.UnescapeString(string(chapterInfo[4]))
		if title == "" {
			title = titleFromIdentity(identity)
		}

		scanlators, _ := this.fetcher().scanlators().AssignIds([]string{"Unknown [" + this.Name() + "]"})

		chapter := Chapter{MarkedRead: source.MarkAsRead}
		chapter.AddScanlation(ChapterScanlation{
			SourceId:    this.id,
			Scanlators:  JoinScanlators(scanlators),
			Version:     1,
			Color:       false,
			Title:       title,
			Language:    ENGLISH_LANG_ID(),
			MetadataURL: url,
			PageLinks:   make([]string, 0),
		})

		identities = append(identities, identity)
		chapters = append(chapters, chapter)
	}

	return identities, chapters, true
}

func (this *kissmanga) chapterDataLinks(url string) []string {
	contents, err := this.fetcher().DownloadData(this.id, url, false)
	if err != nil {
		panic(err)
	}

	pageLinksBytes := kissmanga_rPageLinks.FindAll(contents, -1)
	pageLinks := make([]string, 0, len(pageLinksBytes))
	for _, pageLink := range pageLinksBytes {
		if proxied := kissmanga_rProxiedImgur.Find(pageLink); len(proxied) == 0 {
			pageLinks = append(pageLinks, string(pageLink))
		} else {
			unescaped, _ := neturl.QueryUnescape(string(proxied))
			pageLinks = append(pageLinks, unescaped)
		}
	}
	return pageLinks
}
