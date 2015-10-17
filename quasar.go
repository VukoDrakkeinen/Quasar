package main

import (
	"flag"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/core"
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
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

	settings, list, vars := initQuasar()

	if err := qml.Run(func() error { return launchGUI(vars) }); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	settings.Save()
	list.SaveToDB() //FIXME: there are still some bugs lurking there

	return
}

func initQuasar() (*core.GlobalSettings, core.ComicList, qmlContextVariables) {
	qlog.Log(qlog.Info, "Loading settings")
	settings, err := core.LoadGlobalSettings()
	if err != nil {
		qlog.Log(qlog.Error, "Loading settings failed!")
		qlog.Log(qlog.Warning, "Falling back on defaults")
		settings = core.NewGlobalSettings()
	}

	qlog.Log(qlog.Info, "Creating proxy models")
	chapterModel := gui.NewComicChapterModel(nil)
	updateModel := gui.NewComicUpdateModel(nil)
	infoModel := gui.NewComicInfoModel(nil)

	qlog.Log(qlog.Info, "Creating Fetcher")
	notify := gui.DefaultNotifyFunc()
	fet := core.NewFetcher(settings, func(work func()) {
		notify(chapterModel, core.Reset, -1, -1, work) //row and count values are unused, hence -1
	})

	qlog.Log(qlog.Info, "Registering plugins")
	fet.RegisterPlugins(core.NewBatoto(), core.NewBakaUpdates())

	qlog.Log(qlog.Info, "Creating comic list")
	list := core.NewComicList(fet, func(ntype core.ViewNotificationType, row, count int, work func()) {
		notify(updateModel, ntype, row, count, work)
	})
	gui.ModelSetGoData(chapterModel, &list)
	gui.ModelSetGoData(updateModel, &list)
	gui.ModelSetGoData(infoModel, &list)

	qlog.Log(qlog.Info, "Loading from DB intiated")
	go func() {
		err = list.LoadFromDB()
		//err = list.LoadFromDB()	//Test consecutive loads
		if err != nil {
			qlog.Log(qlog.Error, err)
			os.Exit(1)
		}

	}()

	qlog.Log(qlog.Info, "Creating Core Connector")
	coreConnector := gui.NewCoreConnector(&list, func(row int, selections [][2]int, work func()) {
		work()
		for _, sel := range selections {
			gui.NotifyViewUpdated(chapterModel, sel[0], sel[1], -1) //[0] = row, [1] = count
		}
		gui.NotifyViewUpdated(updateModel, row, 1, -1)
	})

	vars := qmlContextVariables{
		{name: "updateModel", ptr: updateModel.InternalPtr()},
		{name: "infoModel", ptr: infoModel.InternalPtr()},
		{name: "chapterModel", ptr: chapterModel.InternalPtr()},
		{name: "quasarCore", ptr: coreConnector, isGoValue: true},
	}
	return settings, list, vars
}

func launchGUI(contextVars qmlContextVariables) error {
	engine := qml.NewEngine()
	engine.On("quit", func() { println("Save to DB here?"); os.Exit(0) })
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
