package gui

// #cgo CPPFLAGS: -I./cpp
// #cgo CXXFLAGS: -std=c++11 -pedantic-errors -Wall -fno-strict-aliasing
// #cgo LDFLAGS: -lstdc++
// #cgo pkg-config: Qt5Core Qt5Widgets Qt5Quick
// #include "cpp/qcapi.h"
import "C"

import (
	"quasar/core"
	"unsafe"
)

//TODO: notify models of changes in Go data

func NewComicInfoModel(list *core.ComicList) (model unsafe.Pointer) {
	return C.newInfoModel(unsafe.Pointer(list))
}

func NewComicUpdateModel(list *core.ComicList) (model unsafe.Pointer) {
	return C.newUpdateModel(unsafe.Pointer(list))
}

func NewComicChapterModel(list *core.ComicList) (model unsafe.Pointer) {
	return C.newChapterModel(unsafe.Pointer(list))
}
