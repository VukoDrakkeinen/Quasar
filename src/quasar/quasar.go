package main

import (
	"flag"
	"fmt"
	"gopkg.in/qml.v1"
	"log"
	"os"
	"quasar/core"
	"quasar/datadir/qlog"
	"quasar/gui"
	"runtime/pprof"
	"time"
)

var _ = time.Kitchen
var _ = os.DevNull

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() { //TODO: messy code, move all that stuff to a dedicated testing suite

	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if err := qml.Run(launchGUI); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	return

	/*globals, _ := core.LoadGlobalSettings()
	qlog.Log(qlog.Info, "Creating Fetcher")
	fet := core.NewFetcher(nil)
	qlog.Log(qlog.Info, "Registering plugins")
	batoto := core.NewBatoto()
	bupdates := core.NewBakaUpdates()
	fet.RegisterPlugin(batoto)
	fet.RegisterPlugin(bupdates)
	qlog.Log(qlog.Info, "Creating Comic")
	comic := core.NewComic(*core.NewIndividualSettings(globals))
	qlog.Log(qlog.Info, "Finding comic URL")
	fet.TestFind(comic, batoto.PluginName(), "Kingdom") //Has lots of data to process, good for testing
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	qlog.Log(qlog.Info, "Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	qlog.Log(qlog.Info, "Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	for i := 0; i < comic.ChapterCount(); i++ {
		chapter, id := comic.GetChapter(i)
		sc0 := chapter.Scanlation(0)
		fmt.Printf("%v %v (%v)\n", id, sc0.Title, sc0.Scanlators)
	}

	//return

	qlog.Log(qlog.Info, "Saving to DB")
	list := core.NewComicList(fet, nil)
	list.AddComics([]*core.Comic{comic})
	//list.ScheduleComicFetches()
	//time.Sleep(5 * time.Second) //Wait for the background tasks to complete
	list.SaveToDB()
	qlog.Log(qlog.Info, "Saved")
	qlog.Log(qlog.Info, "Loading from DB")
	err := list.LoadFromDB()
	if err != nil {
		qlog.Log(qlog.Error, err)
		return
	}
	qlog.Log(qlog.Info, "Loaded")

	return

	fmt.Println("\nDownloading Page Links for Chapter:0 Scanlation:0")
	fet.DownloadPageLinksFor(comic, 0, 0)
	chapter, id := comic.GetChapter(0)
	fmt.Println(id, chapter)//*/
}

var dontGC *core.ComicList //TODO: It's a tra- I mean, a HACK! Remove it!

func launchGUI() error { //TODO: move some things out of GUI thread
	qml.RegisterTypes("QuasarGUI", 1, 0, []qml.TypeSpec{
		{Init: gui.InitSplitDurationValidator},
	})

	engine := qml.NewEngine()
	context := engine.Context()

	qlog.Log(qlog.Info, "Loading settings")
	settings, err := core.LoadGlobalSettings()
	if err != nil {
		qlog.Log(qlog.Error, "Loading settings failed!")
		qlog.Log(qlog.Warning, "Falling back on defaults")
		settings = core.NewGlobalSettings()
	}
	//fmt.Printf("%#v\n", settings)

	qlog.Log(qlog.Info, "Creating proxy models")
	chapterModel := gui.NewComicChapterModel(nil)
	updateModel := gui.NewComicUpdateModel(nil)
	infoModel := gui.NewComicInfoModel(nil)

	qlog.Log(qlog.Info, "Creating Fetcher")
	fet := core.NewFetcher(settings, func(work func()) {
		//println("Notifying chapter model - reset")
		notify := gui.DefaultNotifyFunc()
		notify(chapterModel, core.Reset, -1, -1, work) //row and count values are unused, hence -1
	})

	qlog.Log(qlog.Info, "Registering plugins")
	fet.RegisterPlugins(core.NewBatoto(), core.NewBakaUpdates())

	qlog.Log(qlog.Info, "Creating comic list")
	list := core.NewComicList(fet, func(ntype core.ViewNotificationType, row, count int, work func()) {
		//println("Notifying updateModel with ntype", ntype, "row", row, "count", count)
		notify := gui.DefaultNotifyFunc()
		notify(updateModel, ntype, row, count, work)
	})
	dontGC = &list
	gui.ModelSetGoData(chapterModel, &list)
	gui.ModelSetGoData(updateModel, &list)
	gui.ModelSetGoData(infoModel, &list)

	qlog.Log(qlog.Info, "Loading from DB")
	err = list.LoadFromDB()
	//err = list.LoadFromDB()	//Test consecutive loads
	if err != nil {
		qlog.Log(qlog.Error, err)
		os.Exit(1)
	}

	qlog.Log(qlog.Info, "Setting QML variables")
	context.SetVar("updateModel", qml.CommonOf(updateModel.InternalPtr(), engine))
	context.SetVar("infoModel", qml.CommonOf(infoModel.InternalPtr(), engine))
	context.SetVar("chapterModel", qml.CommonOf(chapterModel.InternalPtr(), engine))
	context.SetVar("quasarCore", gui.NewCoreConnector(&list))

	qlog.Log(qlog.Info, "Launching GUI")
	control, err := engine.LoadFile("/home/vuko/Projects/GoLang/Quasar/src/quasar/gui/qml/main.qml") //TODO: load from resources
	if err != nil {
		return err
	}
	window := control.CreateWindow(nil)

	window.Show()
	window.Wait()
	//settings.Save()	//TODO: fix this (seems to save the default values)
	return nil
}
