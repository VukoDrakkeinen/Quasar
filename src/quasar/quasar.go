package main

import (
	"fmt"
	"gopkg.in/qml.v1"
	//"os"
	"quasar/gui"
	"quasar/redshift"
	"time"
)

func main() {
	/*
		if err := qml.Run(launchGUI); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}

		return//*/

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
	fet.TestFind(comic, batoto.PluginName(), "Kingdom")
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	fmt.Println("Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	fmt.Println(comic.Info)
	fmt.Println("Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	for _, i := range []int{0, 1, 62} {
		chapter, id := comic.GetChapter(i)
		fmt.Printf("%v %+v\n", id, chapter)
	}

	//return
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
	fet.TestFind(comic, batoto.PluginName(), "Kingdom")
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	fmt.Println("Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	fmt.Println(comic.Info)
	fmt.Println("Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	list := redshift.NewComicList(fet)
	list.AddComics([]*redshift.Comic{comic})
	list.ScheduleComicFetches()
	fmt.Println("Waiting 5 seconds for background tasks (cross-thread synchronization not done yet!)...")
	time.Sleep(5 * time.Second)

	modelCommon := qml.CommonOf(gui.NewModel(list), engine)
	//modelCommon := qml.CommonOf(gui.NewDummyModel(), engine)
	context.SetVar("comicListModel", modelCommon)
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
