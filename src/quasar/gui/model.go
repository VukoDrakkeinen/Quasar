package gui

// #cgo CPPFLAGS: -I./cpp
// #cgo CXXFLAGS: -std=c++11 -pedantic-errors -Wall -fno-strict-aliasing
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
// #include "cpp/qcapi.h"
import "C"

import (
	"fmt"
	"math/rand"
	"quasar/redshift"
	"strconv"
	"time"
	"unsafe"
)

type QComicStatus int8

const (
	QNoUpdates QComicStatus = iota
	QUpdating
	QNewChapters
	QError
)

type QInfoRow struct {
	title     string
	chapTotal int
	chapRead  int
	updated   int64
	progress  int
	status    QComicStatus
}
type QList unsafe.Pointer

func NewDummyData() QList {
	gen := rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
	var infoRows []QInfoRow
	for i := 0; i < 10; i++ {
		r := gen.Intn(9948)
		row := QInfoRow{
			title:     "Title " + strconv.FormatInt(int64(r), 10),
			chapTotal: r,
			chapRead:  r / 2,
			updated:   time.Now().Add(-time.Duration(r) * time.Minute * time.Duration(3)).Unix(),
			progress:  r % 101,
			status:    QComicStatus(r % (int(QError) + 1)),
		}
		if row.status == QNoUpdates {
			row.chapRead = row.chapTotal
		}
		infoRows = append(infoRows, row)
	}
	var Integer int
	qlist := C.newList(
		unsafe.Pointer(&infoRows[0]), C.int(unsafe.Sizeof(infoRows[0])), C.int(len(infoRows)),
		C.int(unsafe.Offsetof(infoRows[0].title)), C.int(unsafe.Offsetof(infoRows[0].chapTotal)), C.int(unsafe.Offsetof(infoRows[0].chapRead)),
		C.int(unsafe.Offsetof(infoRows[0].updated)), C.int(unsafe.Offsetof(infoRows[0].progress)), C.int(unsafe.Offsetof(infoRows[0].status)),
		C.int(unsafe.Sizeof(&Integer)), C.int(unsafe.Sizeof(Integer)),
	)
	return QList(qlist)
}

func NewDummyModel() unsafe.Pointer {
	var model unsafe.Pointer = C.newModel(unsafe.Pointer(NewDummyData()))
	//fmt.Println(model)
	return model
}

func NewModel(list redshift.ComicList) unsafe.Pointer {
	var infoRows []QInfoRow
	for i, comic := range list.Hack_Comics() {
		info := comic.Info()
		row := QInfoRow{
			title:     info.Title,
			chapTotal: comic.ChapterCount(),
			chapRead:  comic.ChaptersReadCount(),
			updated:   list.ComicLastUpdated(i).Unix(),
			progress:  100,          //TODO
			status:    QNewChapters, //TODO
		}
		infoRows = append(infoRows, row)
	}
	var Integer int
	qlist := C.newList(
		unsafe.Pointer(&infoRows[0]), C.int(unsafe.Sizeof(infoRows[0])), C.int(len(infoRows)),
		C.int(unsafe.Offsetof(infoRows[0].title)), C.int(unsafe.Offsetof(infoRows[0].chapTotal)), C.int(unsafe.Offsetof(infoRows[0].chapRead)),
		C.int(unsafe.Offsetof(infoRows[0].updated)), C.int(unsafe.Offsetof(infoRows[0].progress)), C.int(unsafe.Offsetof(infoRows[0].status)),
		C.int(unsafe.Sizeof(&Integer)), C.int(unsafe.Sizeof(Integer)),
	)
	var model unsafe.Pointer = C.newModel(unsafe.Pointer(qlist))
	fmt.Println("Model:", model)
	return model
}
