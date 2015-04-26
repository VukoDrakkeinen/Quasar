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
	var slice []QInfoRow
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
		slice = append(slice, row)
	}
	var Integer int
	qlist := C.newList(
		unsafe.Pointer(&slice[0]), C.int(unsafe.Sizeof(slice[0])), C.int(len(slice)),
		C.int(unsafe.Offsetof(slice[0].title)), C.int(unsafe.Offsetof(slice[0].chapTotal)), C.int(unsafe.Offsetof(slice[0].chapRead)),
		C.int(unsafe.Offsetof(slice[0].updated)), C.int(unsafe.Offsetof(slice[0].progress)), C.int(unsafe.Offsetof(slice[0].status)),
		C.int(unsafe.Sizeof(&Integer)), C.int(unsafe.Sizeof(Integer)),
	)
	return QList(qlist)
}

func NewModel( /*list redshift.ComicList*/ ) unsafe.Pointer {
	/*
		for _, comic := range list {
			info := comic.Info
			info.Title
			chapTotal := comic.ChapterCount()
			chapRead := comic.ChaptersReadCount()
			updated := time.Now()
			progress := 30
			status := QNoUpdates
			if chapRead < chapTotal {
				status = QNewChapters
			}
		}//*/
	qlist := NewDummyData()
	var model unsafe.Pointer = C.newModel(unsafe.Pointer(qlist))
	//var model string = "asldjakld"
	fmt.Println(model)
	return model
}
