package main

import (
	"fmt"
	///"gopkg.in/qml.v1"
	///"os"
	"quasar/redshift"
)

func main() {
	///if err := qml.Run(launchGUI); err != nil {
	///	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	///	os.Exit(1)
	///}

	///return

	globals, _ := redshift.LoadGlobalSettings()
	fmt.Println("Creating Fetcher")
	fet := redshift.NewFetcher(nil)
	fmt.Println("Registering plugins")
	batoto := redshift.NewBatoto()
	bupdates := redshift.NewBakaUpdates()
	fet.RegisterPlugin(batoto)
	fet.RegisterPlugin(bupdates)
	fmt.Println("Creating Comic")
	comic := redshift.NewComic()
	comic.Settings = *redshift.NewIndividualSettings(globals)
	fmt.Println("Finding comic URL")
	fet.TestFind(comic, batoto.PluginName(), "Kingdom")
	fet.TestFind(comic, bupdates.PluginName(), "Kingdom")
	/*fmt.Println("Adding UpdateSource")
	comic.AddSource(redshift.UpdateSource{
		PluginName: redshift.FetcherPluginName("BATOTO"),
		//URL:        "http://bato.to/comic/_/comics/guilty-crown-r2323", //lel
		//URL:        "http://bato.to/comic/_/comics/kimi-no-iru-machi-r29",
		URL:        "http://bato.to/comic/_/comics/kingdom-r642",
		MarkAsRead: false,
	})	//*/
	fmt.Println("Downloading ComicInfo")
	fet.DownloadComicInfoFor(comic)
	fmt.Println(comic.Info)
	fmt.Println("Downloading Chapter List")
	fet.DownloadChapterListFor(comic)
	for _, i := range []int{0, 1, 62} {
		chapter, id := comic.GetChapter(i)
		fmt.Printf("%v %+v\n", id, chapter)
	}

	return

	fmt.Println("Saving to DB")
	var list redshift.ComicList
	list = append(list, comic)
	list.SaveToDB()
	fmt.Println("Saved")
	fmt.Println("Loading from DB")
	var err error
	list, err = redshift.LoadComicList()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Loaded")

	return

	fmt.Println("\nDownloading Page Links for Chapter:0 Scanlation:0")
	fet.DownloadPageLinksFor(comic, 0, 0)
	chapter, id := comic.GetChapter(0)
	fmt.Println(id, chapter)
}

///func launchGUI() error {
///	engine := qml.NewEngine()
///
///	controls, err := engine.LoadFile("/home/vuko/Projects/QML/Fullerene-UI/Fullerene-UI.qml")
///	if err != nil {
///		return err
///	}
///	window := controls.CreateWindow(nil)
///	var settings *redshift.GlobalSettings
///	go func() {
///		settings = redshift.LoadGlobalSettings()
///		chooser := window.Object("optsWindow").Object("chooser")
///		chooser.Call("setValues", int(settings.DefaultUpdateNotificationMode), settings.DefaultAccumulativeModeCount, nil) //TODO: duration
///	}()
//go initializeData(window)

//wybieracz := window.Object("optsWindow").Object("chooser")
//wybieracz.On("componentCompleted", func() { fmt.Println("Hello func") })
//wybieracz.Call("setValues", int(redshift.Delayed), 88, nil)

///	window.Show()
///	window.Wait()
///	settings.Save()
///	return nil
///}
