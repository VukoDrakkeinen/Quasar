package gui

// #cgo CPPFLAGS: -I./cpp
// #cgo CXXFLAGS: -std=c++11 -pedantic-errors -Wall -fno-strict-aliasing
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
// #include "cpp/qcapi.h"
import "C"

import (
	"fmt"
	"quasar/redshift"
	. "quasar/redshift/idsdict"
	"reflect"
	"sort"
	"unsafe"
)

type QComicStatus int8

const (
	QNoUpdates QComicStatus = iota
	QUpdating
	QNewChapters
	QError
)

type QUpdateInfoRow struct {
	title     string
	chapTotal int
	chapRead  int
	updated   int64
	progress  int
	status    QComicStatus
}

func NewModel(list redshift.ComicList) (model unsafe.Pointer) {
	var updateInfoRows []QUpdateInfoRow
	for i, comic := range list.Hack_Comics() {
		info := comic.Info()
		row := QUpdateInfoRow{
			title:     info.Title,
			chapTotal: comic.ChapterCount(),
			chapRead:  comic.ChaptersReadCount(),
			updated:   list.ComicLastUpdated(i).Unix(),
			progress:  100,          //TODO
			status:    QNewChapters, //TODO
		}
		updateInfoRows = append(updateInfoRows, row)
	}

	var elem QUpdateInfoRow
	qlist := C.newList(
		arrayPtr(updateInfoRows), C.int(unsafe.Sizeof(elem)), C.int(len(updateInfoRows)),
		C.int(unsafe.Offsetof(elem.title)), C.int(unsafe.Offsetof(elem.chapTotal)), C.int(unsafe.Offsetof(elem.chapRead)),
		C.int(unsafe.Offsetof(elem.updated)), C.int(unsafe.Offsetof(elem.progress)), C.int(unsafe.Offsetof(elem.status)),
	)
	model = C.newModel(unsafe.Pointer(qlist))
	fmt.Println("Model:", model)
	return model
}

type CoComicInfo struct {
	Titles     []string
	Genres     []ComicGenreId
	Categories []ComicTagId
}

func NewComicInfoModel(list redshift.ComicList) (model unsafe.Pointer) {
	var infos []redshift.ComicInfo
	var coInfos []CoComicInfo
	for _, comic := range list.Hack_Comics() {
		info := comic.Info()
		infos = append(infos, info)
		var altTitles []string
		var genres []ComicGenreId
		var tags []ComicTagId
		for altTitle := range info.AltTitles {
			altTitles = append(altTitles, altTitle)
		}
		for genre := range info.Genres {
			genres = append(genres, genre)
		}
		for tag := range info.Categories {
			tags = append(tags, tag)
		}
		sort.Strings(altTitles)
		var compileTimeTypeCheck int
		var failurePoint Id
		compileTimeTypeCheck = int(failurePoint) //Will fail if Id's underlying type will be ever changed from int
		failurePoint = Id(compileTimeTypeCheck)  //I'm just too lazy to write sortable SomethingIdSlice structs :3
		sort.Ints(*(*[]int)(unsafe.Pointer(&genres)))
		sort.Ints(*(*[]int)(unsafe.Pointer(&tags)))
		coinfo := CoComicInfo{
			Titles:     altTitles,
			Genres:     genres,
			Categories: tags,
		}
		coInfos = append(coInfos, coinfo)
	}

	var elem1 redshift.ComicInfo
	var elem2 CoComicInfo
	offsets := [...]uintptr{
		unsafe.Offsetof(elem1.Title),
		unsafe.Offsetof(elem1.Authors),
		unsafe.Offsetof(elem1.Artists),
		unsafe.Offsetof(elem1.Type),
		unsafe.Offsetof(elem1.Status),
		unsafe.Offsetof(elem1.ScanlationStatus),
		unsafe.Offsetof(elem1.Description),
		unsafe.Offsetof(elem1.Rating),
		unsafe.Offsetof(elem1.Mature),
		unsafe.Offsetof(elem1.ThumbnailFilename),
		unsafe.Offsetof(elem2.Titles),
		unsafe.Offsetof(elem2.Genres),
		unsafe.Offsetof(elem2.Categories),
	}

	qlist := C.newComicInfoList(
		arrayPtr(infos), arrayPtr(coInfos),
		C.int(len(infos)), C.int(unsafe.Sizeof(elem1)), C.int(unsafe.Sizeof(elem2)),
		unsafe.Pointer(&offsets),
	)
	model = C.newInfoModel(unsafe.Pointer(qlist))
	fmt.Println("IModel:", model)
	return model
}

func arrayPtr(slice interface{}) (internal unsafe.Pointer) {
	val := reflect.ValueOf(slice)
	if val.Kind() == reflect.Slice {
		return unsafe.Pointer(val.Pointer())
	}
	panic("arrayPtr: invalid type, expected Array or Slice, got " + val.Kind().String())
}
