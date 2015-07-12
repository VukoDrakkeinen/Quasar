package main

import (
	"flag"
	"fmt"
	"gopkg.in/qml.v1"
	"log"
	"os"
	"quasar/gui"
	"quasar/redshift"
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

	globals, _ := redshift.LoadGlobalSettings()
	fmt.Println("Creating Fetcher")
	fet := redshift.NewFetcher(nil)
	fmt.Println("Registering plugins")
	batoto := redshift.NewBatoto()
	bupdates := redshift.NewBakaUpdates()
	fet.RegisterPlugin(batoto)
	fet.RegisterPlugin(bupdates)
	fmt.Println("Creating Comic")
	comic := redshift.NewComic(*redshift.NewIndividualSettings(globals))
	fmt.Println("Finding comic URL")
	fet.TestFind(comic, batoto.PluginName(), "Kingdom") //Has lots of data to process, good for testing
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	fmt.Println("Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	fmt.Println(comic.Info)
	fmt.Println("Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	for i := 0; i < comic.ChapterCount(); i++ {
		chapter, id := comic.GetChapter(i)
		sc0 := chapter.Scanlation(0)
		fmt.Printf("%v %v (%v)\n", id, sc0.Title, sc0.Scanlators)
	}

	return
	/*
		fmt.Println("Saving to DB")
		list := redshift.NewComicList(fet)
		list.AddComics([]*redshift.Comic{comic})
		list.ScheduleComicFetches()
		time.Sleep(5 * time.Second) //Wait for the background tasks to complete
		list.SaveToDB()
		fmt.Println("Saved")
		fmt.Println("Loading from DB")
		err := list.LoadFromDB()
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println("Loaded")

		return
	*/
	fmt.Println("\nDownloading Page Links for Chapter:0 Scanlation:0")
	fet.DownloadPageLinksFor(comic, 0, 0)
	chapter, id := comic.GetChapter(0)
	fmt.Println(id, chapter)
}

var dontGC *redshift.ComicList //TODO: It's a tra- I mean, a HACK! Remove it!

func launchGUI() error {
	qml.RegisterTypes("QuasarGUI", 1, 0, []qml.TypeSpec{
		{Init: gui.InitSplitDurationValidator},
	})

	engine := qml.NewEngine()
	context := engine.Context()

	globals, _ := redshift.LoadGlobalSettings()
	fmt.Println("Creating Fetcher")
	fet := redshift.NewFetcher(nil)
	fmt.Println("Registering plugins")
	batoto := redshift.NewBatoto()
	bupdates := redshift.NewBakaUpdates()
	fet.RegisterPlugin(batoto)
	fet.RegisterPlugin(bupdates)
	fmt.Println("Creating Comic")
	comic := redshift.NewComic(*redshift.NewIndividualSettings(globals))
	fmt.Println("Finding comic URL")
	fet.TestFind(comic, batoto.PluginName(), "Kingdom") //Has lots of data to process, good for testing
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	fmt.Println("Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	fmt.Println("Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	list := redshift.NewComicList(fet)
	list.AddComics([]*redshift.Comic{comic})
	dontGC = &list
	//list.ScheduleComicFetches()
	//fmt.Println("Waiting 5 seconds for background tasks (model notification not done yet!)...")
	//time.Sleep(5 * time.Second)

	fmt.Println("Crash nao!")
	updatemodelCommon := qml.CommonOf(gui.NewComicUpdateModel(&list), engine)
	infomodelCommon := qml.CommonOf(gui.NewComicInfoModel(&list), engine)
	chaptermodelCommon := qml.CommonOf(gui.NewComicChapterModel(&list), engine)
	fmt.Println("Crash niet")
	//modelCommon := qml.CommonOf(gui.NewDummyModel(), engine)
	context.SetVar("updateModel", updatemodelCommon)
	context.SetVar("infoModel", infomodelCommon)
	context.SetVar("chapterModel", chaptermodelCommon)
	//context.SetVar("notifModeChooser", 0)

	controls, err := engine.LoadFile("/home/vuko/Projects/GoLang/Quasar/src/quasar/gui/qml/main.qml")
	if err != nil {
		return err
	}
	window := controls.CreateWindow(nil)
	//var settings *redshift.GlobalSettings
	/*go func() {
		chooser := window.ObjectByName("notifModeChooser")
		chooser.Call("setValues", int(settings.DefaultNotificationMode), settings.DefaultAccumulativeModeCount, nil) //TODO: duration
	}()//*/

	window.Show()
	window.Wait()
	//settings.Save()
	return nil
}
