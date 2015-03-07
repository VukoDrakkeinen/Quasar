package redshift

import (
	"quasar/qutils"
	. "quasar/redshift/idbase"
)

type scanlationIndex int
type Chapter2 struct {
	parent      *Comic
	scanlations []ChapterScanlation
	mapping     map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex
	usedPlugins []FetcherPluginName
	AlreadyRead bool
}

type ChapterScanlation struct {
	Title      string
	Language   LangId
	Scanlators JointScanlatorIds
	PluginName FetcherPluginName
	URL        string
	PageLinks  []string
}

func (this *Chapter2) Scanlation(index int) ChapterScanlation {
	this.initialize()
	if this.parent == nil { //We have no parent, so we can't access priority lists for plugins and scanlators
		return this.scanlations[index]
	} else {
		pluginName, scanlators := this.indexToPath(index)
		return this.scanlations[this.mapping[pluginName][scanlators]]
	}
}

func (this *Chapter2) ScanlationsCount() int {
	return len(this.scanlations)
}

func (this *Chapter2) MergeWith(another *Chapter2) *Chapter2 {
	this.AlreadyRead = another.AlreadyRead || this.AlreadyRead
	for _, scanlation := range another.scanlations {
		this.AddScanlation(scanlation)
	}
	return this
}

func (this *Chapter2) AddScanlation(scanlation ChapterScanlation) (replaced bool) {
	this.initialize()
	if mapped, pluginExists := this.mapping[scanlation.PluginName]; pluginExists {
		if index, scanlatorsExists := mapped[scanlation.Scanlators]; scanlatorsExists {
			this.scanlations[index] = scanlation
			return true
		}
	} else {
		this.usedPlugins = append(this.usedPlugins, scanlation.PluginName)
	}

	this.mapping[scanlation.PluginName] = make(map[JointScanlatorIds]scanlationIndex)
	this.mapping[scanlation.PluginName][scanlation.Scanlators] = scanlationIndex(len(this.scanlations))
	this.scanlations = append(this.scanlations, scanlation)
	return false
}

func (this *Chapter2) RemoveScanlation(index int) {
	this.initialize()
	pluginName, scanlators := this.indexToPath(index)
	realIndex := this.mapping[pluginName][scanlators]

	this.scanlations = append(this.scanlations[:realIndex], this.scanlations[realIndex+1:]...)
	delete(this.mapping[pluginName], scanlators)

	if len(this.mapping[pluginName]) == 0 {
		delete(this.mapping, pluginName)
		deletionIndex, _ := qutils.IndexOf(this.usedPlugins, pluginName)
		this.usedPlugins = append(this.usedPlugins[:deletionIndex], this.usedPlugins[deletionIndex+1:]...)
	}
}

func (this *Chapter2) RemoveScanlationsForPlugin(pluginName FetcherPluginName) {
	this.initialize()
	for _, realIndex := range this.mapping[pluginName] {
		this.scanlations = append(this.scanlations[:realIndex], this.scanlations[realIndex+1:]...)
	}
	delete(this.mapping, pluginName)
	deletionIndex, _ := qutils.IndexOf(this.usedPlugins, pluginName)
	this.usedPlugins = append(this.usedPlugins[:deletionIndex], this.usedPlugins[deletionIndex+1:]...)
}

func (this *Chapter2) Scanlators() (ret []JointScanlatorIds) {
	this.initialize()
	if this.parent != nil {
		for _, pluginName := range this.usedPlugins {
			perPlugin := this.mapping[pluginName]
			for _, scanlator := range this.parent.scanlatorPriority {
				if _, exists := perPlugin[scanlator]; exists {
					ret = append(ret, scanlator)
				}
			}
		}
	} else {
		for _, scanlation := range this.scanlations {
			ret = append(ret, scanlation.Scanlators)
		}
	}
	return
}

func (this *Chapter2) SetParent(comic *Comic) {
	this.initialize()
	this.parent = comic
}

func (this *Chapter2) indexToPath(index int) (FetcherPluginName, JointScanlatorIds) {
	var pluginNames []FetcherPluginName          //Create a set of plugin names with prioritized ones at the beginning
	for _, source := range this.parent.sources { //Add prioritized plugin names
		if _, exists := this.mapping[source.PluginName]; exists {
			pluginNames = append(pluginNames, source.PluginName)
		}
	}
	for _, pluginName := range this.usedPlugins { //Add the rest
		if _, exists := this.parent.sourceIdxByPlugin[pluginName]; !exists {
			pluginNames = append(pluginNames, pluginName)
		}
	}

	var pluginName FetcherPluginName
	for _, pluginName = range pluginNames { //Absolute index => relative index
		jointsPerPlugin := len(this.mapping[pluginName])
		if index >= jointsPerPlugin {
			index -= (jointsPerPlugin - 1)
		} else {
			break
		}
	}

	scanlatorSet := this.mapping[pluginName]
	var scanlators []JointScanlatorIds
	for _, scanlator := range this.parent.scanlatorPriority { //Create a set of this chapter's scanlators (prioritized first)
		if _, exists := scanlatorSet[scanlator]; exists {
			scanlators = append(scanlators, scanlator)
		}
	}

	return pluginName, scanlators[index]
}

func (this *Chapter2) initialize() {
	if this.mapping == nil {
		this.mapping = make(map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex)
	}
}
