package gui

// #include "qcapi.h"
import "C"

import (
	"github.com/Quasar/core"
	"gopkg.in/qml.v1"
	"sync"
	"unsafe"
)

func NewComicInfoModel(list *core.ComicList) qtProxyModel {
	return qtProxyModel{ptr: C.newInfoModel(unsafe.Pointer(list))}
}

func NewComicUpdateModel(list *core.ComicList) qtProxyModel {
	return qtProxyModel{ptr: C.newUpdateModel(unsafe.Pointer(list))}
}

func NewComicChapterModel(list *core.ComicList) qtProxyModel {
	return qtProxyModel{ptr: C.newChapterModel(unsafe.Pointer(list))}
}

func ModelSetGoData(model qtProxyModel, list *core.ComicList) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() { //run in GUI thread
		C.modelSetGoData(model.ptr, unsafe.Pointer(list))
	})
}

func NotifyViewInsertStart(model qtProxyModel, row, count int) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelInsertStart(model.ptr, C.int(row), C.int(count))
	})
}

func NotifyViewInsertEnd(model qtProxyModel) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelInsertEnd(model.ptr)
	})
}

func NotifyViewRemoveStart(model qtProxyModel, row, count int) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelRemoveStart(model.ptr, C.int(row), C.int(count))
	})
}

func NotifyViewRemoveEnd(model qtProxyModel) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelRemoveEnd(model.ptr)
	})
}

func NotifyViewResetStart(model qtProxyModel) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelResetStart(model.ptr)
	})
}

func NotifyViewResetEnd(model qtProxyModel) {
	model.lock.Lock()
	defer model.lock.Unlock()
	qml.RunMain(func() {
		C.notifyModelResetEnd(model.ptr)
	})
}

type qtProxyModel struct {
	ptr  unsafe.Pointer
	lock sync.Mutex
}

func (this *qtProxyModel) InternalPtr() unsafe.Pointer {
	return this.ptr
}

//work() function is provided by the model and must be executed between notification calls
type NotifyViewFunc func(ntype core.ViewNotificationType, row, count int, work func())

type defaultNotifyViewFunc func(model qtProxyModel, ntype core.ViewNotificationType, row, count int, work func())

func DefaultNotifyFunc() defaultNotifyViewFunc {
	return func(model qtProxyModel, ntype core.ViewNotificationType, row, count int, work func()) {
		switch ntype {
		case core.Insert:
			func() {
				NotifyViewInsertStart(model, row, count)
				defer NotifyViewInsertEnd(model)
				work()
			}()
		case core.Remove:
			func() {
				NotifyViewRemoveStart(model, row, count)
				defer NotifyViewRemoveEnd(model)
				work()
			}()
		case core.Reset:
			func() {
				NotifyViewResetStart(model)
				defer NotifyViewResetEnd(model)
				work()
			}()
		}
	}
}
