package gui

// #include "qcapi.h"
import "C"

import (
	"github.com/VukoDrakkeinen/Quasar/core"
	"gopkg.in/qml.v1"
	//"sync"
	"unsafe"
)

func NewComicInfoModel(list *core.ComicList) qtProxyModel {
	var ptr unsafe.Pointer
	qml.RunMain(func() { //create in GUI thread
		ptr = C.newInfoModel(unsafe.Pointer(list))
	})
	return qtProxyModel{ptr: ptr}
}

func NewComicUpdateModel(list *core.ComicList) qtProxyModel {
	var ptr, filter unsafe.Pointer
	qml.RunMain(func() {
		ptr = C.newUpdateModel(unsafe.Pointer(list))
		filter = C.wrapModel(ptr)
	})
	return qtProxyModel{ptr: ptr, filter: filter}
}

func NewComicChapterModel(list *core.ComicList) qtProxyModel {
	var ptr unsafe.Pointer
	qml.RunMain(func() {
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
	qml.RunMain(func() { //run in GUI thread
		C.modelSetGoData(model.ptr, unsafe.Pointer(list))
	})
}

func (model qtProxyModel) NotifyViewInsertStart(row, count int) {
	qml.RunMain(func() {
		C.notifyModelInsertStart(model.ptr, C.int(row), C.int(count))
	})
}

func (model qtProxyModel) NotifyViewInsertEnd() {
	qml.RunMain(func() {
		C.notifyModelInsertEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewRemoveStart(row, count int) {
	qml.RunMain(func() {
		C.notifyModelRemoveStart(model.ptr, C.int(row), C.int(count))
	})
}

func (model qtProxyModel) NotifyViewRemoveEnd() {
	qml.RunMain(func() {
		C.notifyModelRemoveEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewResetStart() {
	qml.RunMain(func() {
		C.notifyModelResetStart(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewResetEnd() {
	qml.RunMain(func() {
		C.notifyModelResetEnd(model.ptr)
	})
}

func (model qtProxyModel) NotifyViewUpdated(row, count, column int) {
	qml.RunMain(func() {
		C.notifyModelDataChanged(model.ptr, C.int(row), C.int(count), C.int(column))
	})
}

//work() function is provided by the model and must be executed in-between notification calls
type NotifyViewFunc func(ntype core.ViewNotificationType, row, count int, work func())

type defaultNotifyViewFunc func(model qtProxyModel, ntype core.ViewNotificationType, row, count int, work func())

func DefaultNotifyFunc() defaultNotifyViewFunc {
	return func(model qtProxyModel, ntype core.ViewNotificationType, row, count int, work func()) {
		switch ntype {
		case core.Insert:
			func() {
				model.NotifyViewInsertStart(row, count)
				defer model.NotifyViewInsertEnd()
				work()
			}()
		case core.Remove:
			func() {
				model.NotifyViewRemoveStart(row, count)
				defer model.NotifyViewRemoveEnd()
				work()
			}()
		case core.Reset:
			func() {
				model.NotifyViewResetStart()
				defer model.NotifyViewResetEnd()
				work()
			}()
		case core.Update:
			work()
			model.NotifyViewUpdated(row, count, -1)
		}
	}
}
