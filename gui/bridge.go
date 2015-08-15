package gui

// #cgo CPPFLAGS: -I./cpp
// #cgo CXXFLAGS: -std=c++11 -pedantic-errors -Wall -fno-strict-aliasing -O2 -pipe
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
//
// #include "cpp/qcapi.h"
import "C"

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"reflect"
	"sort"
	"sync"
	"unsafe"
)

var dontGC = make(map[unsafe.Pointer]struct{})
var lock sync.Mutex

func init() {
	C.registerQMLTypes()

	var ptr unsafe.Pointer

	ptr = offsets(&comicInfoBridged{})
	disableGcFor(ptr)
	C.go_Offsets_ComicInfo = ptr

	ptr = offsets(&core.ChapterScanlation{})
	disableGcFor(ptr)
	C.go_Offsets_Scanlation = ptr

	ptr = offsets(&updateInfoBridged{})
	disableGcFor(ptr)
	C.go_Offsets_UpdateInfo = ptr
}

//export go_collectGarbage
func go_collectGarbage(ptr unsafe.Pointer) {
	lock.Lock()
	defer lock.Unlock()
	delete(dontGC, ptr)
}

///Ids

//export go_authorNameById
func go_authorNameById(id int) *C.char {
	authorId := *(*idsdict.AuthorId)(unsafe.Pointer(&id))
	return C.CString(idsdict.Authors.NameOf(authorId))
}

//export go_artistNameById
func go_artistNameById(id int) *C.char {
	artistId := *(*idsdict.ArtistId)(unsafe.Pointer(&id))
	return C.CString(idsdict.Artists.NameOf(artistId))
}

//export go_genreNameById
func go_genreNameById(id int) *C.char {
	comicGenreId := *(*idsdict.ComicGenreId)(unsafe.Pointer(&id))
	return C.CString(idsdict.ComicGenres.NameOf(comicGenreId))
}

//export go_categoryNameById
func go_categoryNameById(id int) *C.char {
	comicTagId := *(*idsdict.ComicTagId)(unsafe.Pointer(&id))
	return C.CString(idsdict.ComicTags.NameOf(comicTagId))
}

//export go_scanlatorNameById
func go_scanlatorNameById(id int) *C.char {
	scanlatorId := *(*idsdict.ScanlatorId)(unsafe.Pointer(&id))
	return C.CString(idsdict.Scanlators.NameOf(scanlatorId))
}

//export go_langNameById
func go_langNameById(id int) *C.char {
	langId := *(*idsdict.LangId)(unsafe.Pointer(&id))
	return C.CString(idsdict.Langs.NameOf(langId))
}

//export go_getThumbnailPath
func go_getThumbnailPath(str *C.char) *C.char {
	return C.CString(qdb.GetThumbnailPath(C.GoString(str)))
}

///ComicList

//export go_ComicList_GetComic
func go_ComicList_GetComic(goComicList unsafe.Pointer, idx C.int) unsafe.Pointer {
	list := (*core.ComicList)(goComicList)
	comic := unsafe.Pointer(list.GetComic(int(idx)))
	disableGcFor(comic)
	return comic
}

//export go_ComicList_Len
func go_ComicList_Len(goComicList unsafe.Pointer) C.int {
	list := (*core.ComicList)(goComicList)
	return C.int(list.Len())
}

//export go_ComicList_ComicLastUpdated
func go_ComicList_ComicLastUpdated(goComicList unsafe.Pointer, idx C.int) int64 {
	list := (*core.ComicList)(goComicList)
	return list.ComicLastUpdated(int(idx)).Unix()
}

type qComicStatus int8

const (
	qNoUpdates qComicStatus = iota
	qUpdating
	qNewChapters
	qError
)

type updateInfoBridged struct {
	title         string
	chaptersCount int
	chaptersRead  int
	updated       int64
	progress      int8
	status        qComicStatus
}

//export go_ComicList_ComicUpdateInfo
func go_ComicList_ComicUpdateInfo(goComicList unsafe.Pointer, idx C.int) unsafe.Pointer {
	list := (*core.ComicList)(goComicList)
	comic := list.GetComic(int(idx))

	bridged := &updateInfoBridged{
		title:         comic.Info().Title,
		chaptersCount: comic.ChapterCount(),
		chaptersRead:  comic.ChaptersReadCount(),
		updated:       list.ComicLastUpdated(int(idx)).Unix(),
		progress:      33,     //TODO
		status:        qError, //TODO
	}
	if list.ComicIsUpdating(int(idx)) {
		bridged.status = qUpdating
	} else if bridged.chaptersCount == bridged.chaptersRead {
		bridged.status = qNoUpdates
	} else {
		bridged.status = qNewChapters
	}

	disableGcFor(unsafe.Pointer(bridged))

	return unsafe.Pointer(bridged)
}

///Comic

//export go_Comic_ChaptersCount
func go_Comic_ChaptersCount(goComic unsafe.Pointer) C.int {
	comic := (*core.Comic)(goComic)
	return C.int(comic.ChapterCount())
}

//export go_Comic_ChaptersReadCount
func go_Comic_ChaptersReadCount(goComic unsafe.Pointer) C.int {
	comic := (*core.Comic)(goComic)
	return C.int(comic.ChaptersReadCount())
}

type comicInfoBridged struct {
	Title             string
	AltTitles         []string
	Authors           []idsdict.AuthorId
	Artists           []idsdict.ArtistId
	Genres            []idsdict.ComicGenreId
	Categories        []idsdict.ComicTagId
	Type              int
	Status            int
	ScanlationStatus  int
	Description       string
	Rating            float32
	Mature            bool
	ThumbnailFilename string
}

//export go_Comic_Info
func go_Comic_Info(goComic unsafe.Pointer) unsafe.Pointer {
	comic := (*core.Comic)(goComic)
	info := comic.Info()

	bridged := &comicInfoBridged{
		Title:             info.Title,
		Authors:           info.Authors,
		Artists:           info.Artists,
		Type:              int(info.Type),
		Status:            int(info.Status),
		ScanlationStatus:  int(info.ScanlationStatus),
		Description:       info.Description,
		Rating:            info.Rating,
		Mature:            info.Mature,
		ThumbnailFilename: info.ThumbnailFilename,
	}
	for altTitle := range info.AltTitles {
		bridged.AltTitles = append(bridged.AltTitles, altTitle)
	}
	for genre := range info.Genres {
		bridged.Genres = append(bridged.Genres, genre)
	}
	for tag := range info.Categories {
		bridged.Categories = append(bridged.Categories, tag)
	}
	sort.Strings(bridged.AltTitles) //TODO?: return unsorted, sort in C++-land?
	var compileTimeTypeCheck int
	var failurePoint idsdict.Id
	compileTimeTypeCheck = int(failurePoint)        //Will fail if Id's underlying type will be ever changed from int
	failurePoint = idsdict.Id(compileTimeTypeCheck) //I'm just too lazy to write sortable SomethingIdSlice structs :3
	sort.Ints(*(*[]int)(unsafe.Pointer(&bridged.Genres)))
	sort.Ints(*(*[]int)(unsafe.Pointer(&bridged.Categories)))

	disableGcFor(unsafe.Pointer(bridged))

	return unsafe.Pointer(bridged)
}

//export go_Comic_GetChapter
func go_Comic_GetChapter(goComic unsafe.Pointer, idx C.int) unsafe.Pointer {
	comic := (*core.Comic)(goComic)
	chapter, _ := comic.GetChapter(int(idx))
	disableGcFor(unsafe.Pointer(&chapter))
	return unsafe.Pointer(&chapter)
}

///Chapter

//export go_Chapter_AlreadyRead
func go_Chapter_AlreadyRead(goChapter unsafe.Pointer) bool {
	chapter := (*core.Chapter)(goChapter)
	return chapter.AlreadyRead
}

//export go_Chapter_ScanlationsCount
func go_Chapter_ScanlationsCount(goChapter unsafe.Pointer) C.int {
	chapter := *(*core.Chapter)(goChapter)
	return C.int(chapter.ScanlationsCount())
}

//export go_Chapter_GetScanlation
func go_Chapter_GetScanlation(goChapter unsafe.Pointer, idx C.int) unsafe.Pointer {
	chapter := (*core.Chapter)(goChapter)
	scanlation := chapter.Scanlation(int(idx))
	disableGcFor(unsafe.Pointer(&scanlation))
	return unsafe.Pointer(&scanlation)
}

//export go_JointScanlators_ToSlice
func go_JointScanlators_ToSlice(goJointScanlators uintptr) uintptr {
	jointScanlators := (*idsdict.JointScanlatorIds)(unsafe.Pointer(goJointScanlators))
	slice := jointScanlators.ToSlice()
	disableGcFor(unsafe.Pointer(&slice))
	return uintptr(unsafe.Pointer(&slice))
}

///internal

func disableGcFor(ptr unsafe.Pointer) {
	lock.Lock()
	defer lock.Unlock()
	dontGC[ptr] = struct{}{}
}

func offsets(dataPtr interface{}) (offsets unsafe.Pointer) {
	typ := reflect.TypeOf(dataPtr).Elem()
	fieldCount := typ.NumField()

	offsetsSlice := make([]uintptr, fieldCount)
	for i := 0; i < fieldCount; i++ {
		offsetsSlice[i] = typ.Field(i).Offset
	}
	return arrayPtr(offsetsSlice)
}

func arrayPtr(slice interface{}) (internal unsafe.Pointer) {
	val := reflect.ValueOf(slice)
	for kind := val.Kind(); kind == reflect.Interface || kind == reflect.Ptr; {
		val = val.Elem()
	}
	if val.Kind() == reflect.Slice {
		return unsafe.Pointer(val.Pointer())
	}
	panic("arrayPtr: invalid type, expected Slice, got " + val.Kind().String())
}
