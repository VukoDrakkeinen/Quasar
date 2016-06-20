package gui

// #cgo CPPFLAGS: -I./cpp
// #cgo CXXFLAGS: -std=c++14 -pedantic-errors -Wall -fno-strict-aliasing -pipe -ggdb
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
//
// #include "cpp/qcapi.h"
import "C"

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/core/idsdict"
	"github.com/VukoDrakkeinen/Quasar/datadir/qdb"
	"github.com/VukoDrakkeinen/Quasar/qutils/hashtype"
	"reflect"
	"sync"
	"unsafe"
)

var preventGC = make(map[unsafe.Pointer]struct{})
var lock sync.Mutex

func init() {
	C.registerQMLTypes()

	var ptr unsafe.Pointer

	ci := core.ComicInfo{}
	ptr = offsets(&ci)
	disableGcFor(ptr)
	C.go_Offsets_ComicInfo = ptr
	C.go_Hash_ComicInfo = C.ulonglong(hashtype.Struct(ci))

	cs := core.ChapterScanlation{}
	ptr = offsets(&cs)
	disableGcFor(ptr)
	C.go_Offsets_Scanlation = ptr
	C.go_Hash_Scanlation = C.ulonglong(hashtype.Struct(cs))

	ib := updateInfoBridged{}
	ptr = offsets(&ib)
	disableGcFor(ptr)
	C.go_Offsets_UpdateInfo = ptr
	C.go_Hash_UpdateInfo = C.ulonglong(hashtype.Struct(ib))

	result := C.assertSyncedHashes()
	switch result {
	case 1:
		panic("ComicInfo struct has changed. Update the C glue code.")
	case 2:
		panic("ChapterScanlation struct has changed. Update the C glue code.")
	case 3:
		panic("UpdateInfo struct has changed. Update the C glue code.")
	}
}

//export go_collectGarbage
func go_collectGarbage(ptr unsafe.Pointer) {
	lock.Lock()
	defer lock.Unlock()
	delete(preventGC, ptr)
}

///Ids

//export go_authorNameById
func go_authorNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	authorId := *(*idsdict.AuthorId)(unsafe.Pointer(&id))
	return C.CString(list.Authors().NameOf(authorId))
}

//export go_artistNameById
func go_artistNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	artistId := *(*idsdict.ArtistId)(unsafe.Pointer(&id))
	return C.CString(list.Artists().NameOf(artistId))
}

//export go_genreNameById
func go_genreNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	comicGenreId := *(*idsdict.ComicGenreId)(unsafe.Pointer(&id))
	return C.CString(list.Genres().NameOf(comicGenreId))
}

//export go_categoryNameById
func go_categoryNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	comicTagId := *(*idsdict.ComicTagId)(unsafe.Pointer(&id))
	return C.CString(list.Tags().NameOf(comicTagId))
}

//export go_scanlatorNameById
func go_scanlatorNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	scanlatorId := *(*idsdict.ScanlatorId)(unsafe.Pointer(&id))
	return C.CString(list.Scanlators().NameOf(scanlatorId))
}

//export go_langNameById
func go_langNameById(goComicList unsafe.Pointer, id int) *C.char {
	list := (*core.ComicList)(goComicList)
	langId := *(*idsdict.LangId)(unsafe.Pointer(&id))
	return C.CString(list.Langs().NameOf(langId))
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
func go_ComicList_ComicLastUpdated(goComicList unsafe.Pointer, idx C.int) int64 { //todo: unused?
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

	info := comic.Info()
	bridged := &updateInfoBridged{
		title:         info.Titles[info.MainTitleIdx],
		chaptersCount: comic.ChaptersCount(),
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

	return C.convertUpdateInfo(unsafe.Pointer(bridged))
}

///Comic

//export go_Comic_ChaptersCount
func go_Comic_ChaptersCount(goComic unsafe.Pointer) C.int {
	comic := (*core.Comic)(goComic)
	return C.int(comic.ChaptersCount())
}

//export go_Comic_ChaptersReadCount
func go_Comic_ChaptersReadCount(goComic unsafe.Pointer) C.int {
	comic := (*core.Comic)(goComic)
	return C.int(comic.ChaptersReadCount())
}

//export go_Comic_Info
func go_Comic_Info(goComic unsafe.Pointer) unsafe.Pointer {
	comic := (*core.Comic)(goComic)
	info := comic.Info()
	return C.convertComicInfo(unsafe.Pointer(&info))
}

//export go_Comic_GetChapter
func go_Comic_GetChapter(goComic unsafe.Pointer, idx C.int) unsafe.Pointer { //todo: return handle?
	comic := (*core.Comic)(goComic)
	chapter, _ := comic.Chapter(int(idx))
	return unsafe.Pointer(chapter)
}

///Chapter

//export go_Chapter_AlreadyRead
func go_Chapter_AlreadyRead(goChapter unsafe.Pointer) bool {
	chapter := (*core.Chapter)(goChapter)
	return chapter.MarkedRead
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
	scanlators := scanlation.Scanlators.Slice()
	return C.convertScanlation(unsafe.Pointer(&scanlation), unsafe.Pointer(&scanlators))
}

///internal

func disableGcFor(ptr unsafe.Pointer) {
	lock.Lock()
	defer lock.Unlock()
	preventGC[ptr] = struct{}{}
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
