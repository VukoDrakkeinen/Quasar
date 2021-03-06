package gui

// #include "qcapi.h"
import "C"

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/qml"
	"unsafe"
)

func NewComicInfoModel(list *core.ComicList) qtProxyModel {
	var ptr unsafe.Pointer
	qml.RunInMain(func() { //create in GUI thread
		ptr = C.newInfoModel(unsafe.Pointer(list))
	})
	return qtProxyModel{ptr: ptr}
}

func NewComicUpdateModel(list *core.ComicList) qtProxyModel {
	var ptr, filter unsafe.Pointer
	qml.RunInMain(func() {
		ptr = C.newUpdateModel(unsafe.Pointer(list))
		filter = C.wrapModel(ptr)
	})
	return qtProxyModel{ptr: ptr, filter: filter}
}

func NewComicChapterModel(list *core.ComicList) qtProxyModel {
	var ptr unsafe.Pointer
	qml.RunInMain(func() {
		ptr = C.newChapterModel(unsafe.Pointer(list))
	})
	return qtProxyModel{ptr: ptr}
}

type qtProxyModel struct {
	ptr    unsafe.Pointer
	filter unsafe.Pointer
}

func (this qtProxyModel) QtPtr() unsafe.Pointer {
	var null unsafe.Pointer
	if this.filter != null { //heh, null
		return this.filter
	}
	return this.ptr
}

func (model *qtProxyModel) SetGoData(list *core.ComicList) {
	qml.RunInMain(func() { //run in GUI thread
		C.modelSetGoData(model.ptr, unsafe.Pointer(list))
	})
}

func (model qtProxyModel) NotifyViewInsertStart(row, count int) {
	qml.RunInMain(func() {
		C.notifyModelInsertStart(model.ptr, C.int(row), C.int(count))
	})
}

func (model qtProxyModel) NotifyViewInsertEnd() {
	qml.RunInMain(func() {
		C.notifyModelInsertEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewRemoveStart(row, count int) {
	qml.RunInMain(func() {
		C.notifyModelRemoveStart(model.ptr, C.int(row), C.int(count))
	})
}

func (model qtProxyModel) NotifyViewRemoveEnd() {
	qml.RunInMain(func() {
		C.notifyModelRemoveEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewResetStart() {
	qml.RunInMain(func() {
		C.notifyModelResetStart(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewResetEnd() {
	qml.RunInMain(func() {
		C.notifyModelResetEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewUpdated(row, count, column int) {
	qml.RunInMain(func() {
		C.notifyModelDataChanged(model.ptr, C.int(row), C.int(count), C.int(column))
	})
}
