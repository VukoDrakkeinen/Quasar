package main

import (
	"fmt"
	"os"
	"unsafe"

	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/gui"
	"github.com/VukoDrakkeinen/Quasar/qutils/cores"
	"github.com/VukoDrakkeinen/qml"
	"github.com/pkg/profile"
)

type qmlContextVars []struct {
	name      string
	ptr       interface{}
	isGoValue bool
}

func main() { //TODO: fix messy code; write some unit tests
	if false {
		defer profile.Start().Stop()
	}
	cores.UseAll()

	saveData, vars := initQuasar()

	if err := qml.Run(func() error { return launchGUI(vars, func() { saveData() }) }); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return
}

func initQuasar() (saveData func() error, vars qmlContextVars) {
	qlog.Log(qlog.Info, "Loading settings")
	settings, err := core.LoadGlobalSettings()
	if err != nil {
		qlog.Log(qlog.Error, "Loading settings failed!")
		qlog.Log(qlog.Warning, "Falling back on defaults")
		settings = core.NewGlobalSettings()
	}

	qlog.Log(qlog.Info, "Creating Fetcher")
	fet := core.NewFetcher(settings)

	qlog.Log(qlog.Info, "Registering plugins")
	fet.RegisterPlugins(core.NewBatoto(), core.NewBakaUpdates(), core.NewKissManga())

	qlog.Log(qlog.Info, "Creating comic list")
	list := core.NewComicList(fet)

	qlog.Log(qlog.Info, "Creating proxy models")
	chapterModel := gui.NewComicChapterModel(list)
	updateModel := gui.NewComicUpdateModel(list)
	infoModel := gui.NewComicInfoModel(list)

	qlog.Log(qlog.Info, "Registering event actions")
	list.On(core.ChapterListAboutToChange).Do(func(...interface{}) {
		chapterModel.NotifyViewResetStart()
	})
	list.On(core.ChapterListChanged).Do(func(...interface{}) {
		chapterModel.NotifyViewResetEnd()
	})

	list.On(core.ComicsAboutToBeAdded).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewInsertStart(row, count)
	})
	list.On(core.ComicsAdded).Do(func(...interface{}) {
		updateModel.NotifyViewInsertEnd()
	})

	list.On(core.ComicsAboutToBeRemoved).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewRemoveStart(row, count)
	})
	list.On(core.ComicsRemoved).Do(func(...interface{}) {
		updateModel.NotifyViewRemoveEnd()
	})

	list.On(core.ComicsUpdateStatusChanged).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewUpdated(row, count, -1)
	})

	qlog.Log(qlog.Info, "Begin DB load")
	go func() {
		err = list.LoadFromDB()
		if err != nil {
			qlog.Log(qlog.Error, err)
			os.Exit(1)
		}

	}()

	qlog.Log(qlog.Info, "Creating Core Connector")
	coreConnector := gui.NewCoreConnector(list)
	coreConnector.On(gui.ChaptersMarked).Do(func(args ...interface{}) {
		row := args[0].(int)
		selections := args[1].([][2]int)
		for _, sel := range selections {
			chapterModel.NotifyViewUpdated(sel[0], sel[1], -1) //[0] = row, [1] = count
		}
		updateModel.NotifyViewUpdated(row, 1, -1)
	})

	qvars := qmlContextVars{
		{name: "updateModel", ptr: updateModel.QtPtr()},
		{name: "infoModel", ptr: infoModel.QtPtr()},
		{name: "chapterModel", ptr: chapterModel.QtPtr()},
		{name: "quasarCore", ptr: coreConnector, isGoValue: true},
	}
	saveDataFunc := func() error {
		settings.Save()
		//list.SaveToDB()
		return nil
	}
	return saveDataFunc, qvars
}

func launchGUI(contextVars qmlContextVars, onQuit func()) error {
	engine := qml.NewEngine()
	context := engine.Context()

	qlog.Log(qlog.Info, "Setting QML variables")
	for _, ctxVar := range contextVars {
		if !ctxVar.isGoValue {
			context.SetVar(ctxVar.name, qml.CommonOf(ctxVar.ptr.(unsafe.Pointer), engine))
		} else {
			context.SetVar(ctxVar.name, ctxVar.ptr)
		}
	}

	qlog.Log(qlog.Info, "Launching GUI")
	control, err := engine.LoadFile("qml/main.qml") //TODO: load from resources
	if err != nil {
		return err
	}
	window := control.CreateWindow(nil)
	engine.On("quit", func() { onQuit(); window.Hide() /*os.Exit(0)*/ })

	window.Show()
	window.Wait()
	//onQuit()

	return nil
}
