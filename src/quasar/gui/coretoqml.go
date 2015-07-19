package gui

import (
	"quasar/core"
	"unsafe"
)

type CoreToQML struct {
	List *core.ComicList
}

func (this *CoreToQML) PluginNames() (names *[]string, humanReadableNames *[]string) {
	pluginNames, hrNames := this.List.Fetcher().Plugins() //TODO: important! sorted data!
	//screw FetcherPluginName -> string conversion, we'll have to convert it into C++ data anyway
	return (*[]string)(unsafe.Pointer(&pluginNames)), &hrNames
}

func (this *CoreToQML) PluginAutodetect(url string) (pluginName string) {
	fetcherPluginName, _ := this.List.Fetcher().PluginNameFromURL(url)
	return *(*string)(unsafe.Pointer(&fetcherPluginName))
}
