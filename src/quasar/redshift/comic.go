package redshift

import (
	"image"
	"math"
	. "quasar/redshift/idbase"
)

type sourceIndex int
type priorityIndex int
type Comic struct {
	Info     ComicInfo
	Settings IndividualSettings

	sourceIdxByPlugin map[FetcherPluginName]sourceIndex //also pluginSet
	sources           []UpdateSource                    //also pluginPriority
	chaptersOrder     ChapterIdentitiesSlice
	chapters          map[ChapterIdentity]Chapter
	scanlatorPriority []JointScanlatorIds
}

func (this *Comic) initialize() *Comic {
	if this.sourceIdxByPlugin == nil {
		this.sourceIdxByPlugin = make(map[FetcherPluginName]sourceIndex)
		this.chapters = make(map[ChapterIdentity]Chapter)
	}
	return this
}

func (this *Comic) AddSource(source UpdateSource) (alreadyAdded bool) {
	this.initialize()
	return this.AddSourceAt(len(this.sources), source)
}

func (this *Comic) AddSourceAt(index int, source UpdateSource) (alreadyAdded bool) {
	this.initialize()
	existingIndex, alreadyAdded := this.sourceIdxByPlugin[source.PluginName]
	if alreadyAdded {
		this.sources[existingIndex] = source //replace
	} else {
		if index < len(this.sources) { //insert
			this.sources = append(this.sources, UpdateSource{}) //grow the slice
			copy(this.sources[index+1:], this.sources[index:])  //move the data we want to after our value by one
			this.sources[index] = source
		} else { //append
			this.sources = append(this.sources, source)
		}
		this.sourceIdxByPlugin[source.PluginName] = sourceIndex(index)
	}
	return
}

func (this *Comic) RemoveSource(source UpdateSource) (success bool) {
	this.initialize()
	index, exists := this.sourceIdxByPlugin[source.PluginName]
	if exists {
		this.sources = append(this.sources[:index], this.sources[index+1:]...)
	}
	return exists
}

func (this *Comic) Sources() []UpdateSource {
	this.initialize()
	ret := make([]UpdateSource, len(this.sources))
	copy(ret, this.sources)
	return ret
}

func (this *Comic) GetSource(pluginName FetcherPluginName) UpdateSource { //TODO: not found -> error?
	this.initialize()
	index := this.sourceIdxByPlugin[pluginName]
	return this.sources[index]
}

func (this *Comic) AddChapter(identity ChapterIdentity, chapter *Chapter) (merged bool) {
	this.initialize()
	existingChapter, merged := this.chapters[identity]
	if merged {
		existingChapter.MergeWith(chapter)
		this.chapters[identity] = existingChapter //TODO: use pointers instead?
	} else {
		chapter.SetParent(this)
		this.chapters[identity] = *chapter
		this.chaptersOrder = this.chaptersOrder.Insert(this.chaptersOrder.vestedIndexOf(identity), identity)
	}
	return
}

func (this *Comic) AddMultipleChapters(identities []ChapterIdentity, chapters []Chapter) {
	this.initialize()
	minLen := int(math.Min(float64(len(identities)), float64(len(chapters))))
	nonexistentSlices := make([][]ChapterIdentity, 0, minLen/2) //Slice of slices of non-existent identities
	startIndex := 0                                             //Starting index of new slice of non-existent identities
	newStart := false                                           //Status of creation of the slice
	for i := 0; i < minLen; i++ {
		identity := identities[i]
		chapter := chapters[i]
		existingChapter, exists := this.chapters[identity]
		if exists {
			existingChapter.MergeWith(&chapter)
			if newStart { //Sequence ended, add newly created slice to the list, set creation status to false
				nonexistentSlices = append(nonexistentSlices, identities[startIndex:i])
				newStart = false
			}
			this.chapters[identity] = existingChapter //TODO: use pointers instead?
		} else {
			chapter.SetParent(this)
			this.chapters[identity] = chapter
			if !newStart { //Sequence started, set starting index, set creation status to true
				startIndex = i
				newStart = true
			}
		}
	}
	if newStart { //Sequence ended
		nonexistentSlices = append(nonexistentSlices, identities[startIndex:])
		newStart = false
	}

	for i := 0; i < len(nonexistentSlices); i++ {
		neSlice := nonexistentSlices[i]
		insertionIndex := int(this.chaptersOrder.vestedIndexOf(neSlice[0]))
		this.chaptersOrder = this.chaptersOrder.InsertMultiple(insertionIndex, neSlice)
	}
}

func (this *Comic) GetChapter(index int) (Chapter, ChapterIdentity) { //FIXME: bounds check?
	this.initialize()
	identity := this.chaptersOrder[index]
	return this.chapters[identity], identity
}

func (this *Comic) ScanlatorsPriority() []JointScanlatorIds {
	this.initialize()
	ret := make([]JointScanlatorIds, len(this.sources))
	copy(ret, this.scanlatorPriority)
	return ret
}

func (this *Comic) SetScanlatorsPriority(priority []JointScanlatorIds) {
	this.initialize()
	this.scanlatorPriority = priority
}

func (this *Comic) ChapterCount() int {
	return len(this.chaptersOrder)
}

type UpdateSource struct {
	PluginName FetcherPluginName
	URL        string
	MarkAsRead bool
}

type ComicInfo struct {
	Title            string
	AltTitles        map[string]struct{}
	Authors          []AuthorId
	Artists          []ArtistId
	Genres           map[ComicGenreId]struct{}
	Categories       map[ComicTagId]struct{}
	Type             comicType
	Status           comicStatus
	ScanlationStatus ScanlationStatus
	Description      string
	Rating           float32
	Mature           bool
	Thumbnail        image.Image
}

func (this *ComicInfo) initialize() *ComicInfo {
	if this.AltTitles == nil {
		this.AltTitles = make(map[string]struct{})
		this.Genres = make(map[ComicGenreId]struct{})
		this.Categories = make(map[ComicTagId]struct{})
	}
	return this
}

func (this *ComicInfo) MergeWith(another *ComicInfo) {
	this.initialize()
	if this.Title == "" {
		this.Title = another.Title
	}

	for altTitle, _ := range another.AltTitles {
		this.AltTitles[altTitle] = struct{}{}
	}

	authorsSet := make(map[AuthorId]struct{})
	for _, author := range this.Authors {
		authorsSet[author] = struct{}{}
	}
	for _, author := range another.Authors {
		if _, exists := authorsSet[author]; !exists {
			this.Authors = append(this.Authors, author)
		}
	}

	artistsSet := make(map[ArtistId]struct{})
	for _, artist := range this.Artists {
		artistsSet[artist] = struct{}{}
	}
	for _, artist := range another.Artists {
		if _, exists := artistsSet[artist]; !exists {
			this.Artists = append(this.Artists, artist)
		}
	}

	for genre, _ := range another.Genres {
		this.Genres[genre] = struct{}{}
	}

	for tag, _ := range another.Categories {
		this.Categories[tag] = struct{}{}
	}

	if this.Type == InvalidComic {
		this.Type = another.Type
	}

	if (this.Status == ComicStatusInvalid) ||
		(this.Status == ComicOngoing && another.Status == ComicOnHiatus) ||
		(this.Status == ComicOnHiatus && another.Status == ComicOngoing) ||
		(another.Status == ComicDiscontinued) ||
		(another.Status == ComicComplete) {
		this.Status = another.Status
	}

	if (this.ScanlationStatus == ScanlationStatusInvalid) ||
		(this.ScanlationStatus == ScanlationOngoing && another.ScanlationStatus == ScanlationOnHiatus) ||
		(this.ScanlationStatus == ScanlationOnHiatus && another.ScanlationStatus == ScanlationOngoing) ||
		(another.ScanlationStatus == ScanlationDropped) ||
		(another.ScanlationStatus == ScanlationInDesperateNeedOfMoreStaff) ||
		(another.ScanlationStatus == ScanlationComplete) {
		this.ScanlationStatus = another.ScanlationStatus
	}

	if this.Description == "" {
		this.Description = another.Description
	}

	if this.Rating == 0 {
		this.Rating = another.Rating
	} else if another.Rating != 0 {
		this.Rating = (this.Rating + another.Rating) / 2
	}

	this.Mature = another.Mature || this.Mature

	if this.Thumbnail == nil {
		this.Thumbnail = another.Thumbnail
	}
}

type comicType int

const (
	InvalidComic comicType = iota
	Manga
	Manhwa
	Manhua
	Western
	Webcomic
	Other
)

type comicStatus int

const (
	ComicStatusInvalid comicStatus = iota
	ComicComplete
	ComicOngoing
	ComicOnHiatus
	ComicDiscontinued
)
