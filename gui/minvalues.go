package gui

import (
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/qml"
)

func init() {
	qml.RegisterTypes("QuasarGUI", 1, 0, []qml.TypeSpec{
		{Init: initSplitDurationValidator, Name: "DurationValidator"},
	})
}

type ValuesValidator struct {
	ValidationFunc func(objs []qml.Object) (valid bool)
	CorrectionFunc func(objs []qml.Object, valid bool)
	children       []qml.Object
}

func (this *ValuesValidator) BindObject(obj qml.Object) {
	this.children = append(this.children, obj)
}

func (this *ValuesValidator) UnbindObject(obj qml.Object) {
	index, err := qutils.IndexOf(this.children, obj)
	if err == nil {
		this.children = append(this.children[:index], this.children[index+1:]...)
		//this.children[len(this.children)-1] = nil
	}
}

func (this *ValuesValidator) Work() {
	defer func() {
		if err := recover(); err != nil {
			qlog.Log(qlog.Warning, "ValuesValidator: provided function panicked:", err)
		}
	}()
	if len(this.children) == 0 {
		return
	}
	valid := this.ValidationFunc(this.children)
	this.CorrectionFunc(this.children, valid)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

func ValidateSplitDuration(objs []qml.Object) (valid bool) {
	for _, obj := range objs[:len(objs)-1] { //skip first (the list is reversed)
		if obj.Int("value") != 0 {
			return true
		}
	}
	return false
}

func CorrectSplitDuration(objs []qml.Object, valid bool) {
	index := len(objs) - 1 //index of first object (the list is reversed)
	if !valid {
		objs[index].Set("minimumValue", 1)
	} else {
		objs[index].Set("minimumValue", 0)
	}
}

func initSplitDurationValidator(v *ValuesValidator, _ qml.Object) {
	v.ValidationFunc = ValidateSplitDuration
	v.CorrectionFunc = CorrectSplitDuration
}
