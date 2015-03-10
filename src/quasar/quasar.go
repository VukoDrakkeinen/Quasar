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
	fmt.Println("Creating Fetcher")
	//fet := redshift.NewFetcher()
	fet := redshift.Fetcher{}
	fmt.Println("Registering plugins")
	fet.RegisterPlugin(redshift.NewBatoto())
	fet.RegisterPlugin(redshift.NewBUpdates())
	fmt.Println("Creating Comic")
	comic := &redshift.Comic{Settings: *redshift.NewIndividualSettings(redshift.LoadGlobalSettings())}
	fmt.Println("Finding comic URL")
	fet.TestFind(comic, redshift.FetcherPluginName("Batoto"), "Kingdom")
	fet.TestFind(comic, redshift.FetcherPluginName("BUpdates"), "Kingdom")
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
	fmt.Println("Saving to DB")
	var list redshift.ComicList
	list = append(list, *comic)
	list.SaveToDB()
	fmt.Println("Saved")
	return
	fmt.Println("\nDownloading Page Links for Chapter0 alt0")
	fet.DownloadPageLinksFor(comic, 0, 0)
	chapter, id := comic.GetChapter(0)
	fmt.Println(id, chapter)
	//fmt.Println(chapter.DataCount())
	//fmt.Println(chapter.Data(0).PageLinks)
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

///func initializeData(obj *qml.Window) {
///	obj.Object("optsWindow").Object("chooser").Call("setValues", int(redshift.Delayed), 88, nil)
///}
