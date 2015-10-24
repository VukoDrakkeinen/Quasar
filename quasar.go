package main

import (
	"flag"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/eventq"
	"github.com/VukoDrakkeinen/Quasar/gui"
	"github.com/VukoDrakkeinen/Quasar/qutils/cores"
	"gopkg.in/qml.v1"
	"log"
	"os"
	"runtime/pprof"
	"unsafe"
)

type qmlContextVariables []struct {
	name      string
	ptr       interface{}
	isGoValue bool
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() { //TODO: fix messy code; write some unit tests
	cores.UseAll()

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	saveData, vars := initQuasar()

	if err := qml.Run(func() error { return launchGUI(vars, func() { saveData() }) }); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return
}

func initQuasar() (saveData func() error, vars qmlContextVariables) {
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
	eventq.On(core.ChapterListAboutToChange).Do(func(...interface{}) {
		chapterModel.NotifyViewResetStart()
	})
	eventq.On(core.ChapterListChanged).Do(func(...interface{}) {
		chapterModel.NotifyViewResetEnd()
	})

	eventq.On(gui.ChaptersMarked).Do(func(args ...interface{}) {
		row := args[0].(int)
		selections := args[1].([][2]int)
		for _, sel := range selections {
			chapterModel.NotifyViewUpdated(sel[0], sel[1], -1) //[0] = row, [1] = count
		}
		updateModel.NotifyViewUpdated(row, 1, -1)
	})

	eventq.On(core.ComicsAboutToBeAdded).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewInsertStart(row, count)
	})
	eventq.On(core.ComicsAdded).Do(func(...interface{}) {
		updateModel.NotifyViewInsertEnd()
	})

	eventq.On(core.ComicsAboutToBeRemoved).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewRemoveStart(row, count)
	})
	eventq.On(core.ComicsRemoved).Do(func(...interface{}) {
		updateModel.NotifyViewRemoveEnd()
	})

	eventq.On(core.ComicsUpdateStatusChanged).Do(func(args ...interface{}) {
		row := args[0].(int)
		count := args[1].(int)
		updateModel.NotifyViewUpdated(row, count, -1)
	})

	qlog.Log(qlog.Info, "Begin DB load")
	go func() {
		err = list.LoadFromDB()
		//err = list.LoadFromDB()	//Test consecutive loads
		if err != nil {
			qlog.Log(qlog.Error, err)
			os.Exit(1)
		}

	}()

	qlog.Log(qlog.Info, "Creating Core Connector")
	coreConnector := gui.NewCoreConnector(list)

	qvars := qmlContextVariables{
		{name: "updateModel", ptr: updateModel.QtPtr()},
		{name: "infoModel", ptr: infoModel.QtPtr()},
		{name: "chapterModel", ptr: chapterModel.QtPtr()},
		{name: "quasarCore", ptr: coreConnector, isGoValue: true},
	}
	saveDataFunc := func() error {
		settings.Save()
		list.SaveToDB() //FIXME: there are still some bugs lurking there
		return nil
	}
	return saveDataFunc, qvars
}

func launchGUI(contextVars qmlContextVariables, onQuit func()) error {
	engine := qml.NewEngine()
	engine.On("quit", func() { /*onQuit();*/ os.Exit(0) })
	context := engine.Context()

	qlog.Log(qlog.Info, "Setting QML variables")
	for _, cVar := range contextVars {
		if !cVar.isGoValue {
			context.SetVar(cVar.name, qml.CommonOf(cVar.ptr.(unsafe.Pointer), engine))
		} else {
			context.SetVar(cVar.name, cVar.ptr)
		}
	}

	qlog.Log(qlog.Info, "Launching GUI")
	control, err := engine.LoadFile("qml/main.qml") //TODO: load from resources
	if err != nil {
		return err
	}
	window := control.CreateWindow(nil)

	window.Show()
	window.Wait()

	return nil
}
