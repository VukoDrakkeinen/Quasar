package redshift

import (
	"quasar/qutils"
	. "quasar/redshift/idbase"
)

const (
	LQ_Modifier byte = 10 * iota
	MQ_Modifier
	HQ_Modifier
)

type scanlationIndex int
type Chapter struct {
	parent      *Comic
	scanlations []ChapterScanlation
	mapping     map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex
	usedPlugins []FetcherPluginName
	AlreadyRead bool

	sqlId int64
}

type ChapterScanlation struct {
	Title      string
	Language   LangId
	Scanlators JointScanlatorIds
	PluginName FetcherPluginName
	URL        string
	PageLinks  []string

	sqlId    int64
	plSQLIds []int64
}

func (this *Chapter) Scanlation(index int) ChapterScanlation {
	this.initialize()
	pluginName, scanlators := this.indexToPath(index)
	return this.scanlations[this.mapping[pluginName][scanlators]]
}

func (this *Chapter) ScanlationsCount() int {
	return len(this.scanlations)
}

func (this *Chapter) MergeWith(another *Chapter) *Chapter {
	this.AlreadyRead = another.AlreadyRead || this.AlreadyRead
	for _, scanlation := range another.scanlations {
		this.AddScanlation(scanlation)
	}
	return this
}

func (this *Chapter) AddScanlation(scanlation ChapterScanlation) (replaced bool) {
	this.initialize()
	if mapped, pluginExists := this.mapping[scanlation.PluginName]; pluginExists {
		if index, jointExists := mapped[scanlation.Scanlators]; jointExists {
			scanlation.sqlId = this.scanlations[index].sqlId //copy sqlId, so SQLInsert will treat new struct as old modified
			this.scanlations[index] = scanlation             //replace
			return true
		}
	} else {
		this.usedPlugins = append(this.usedPlugins, scanlation.PluginName)
	}

	if this.mapping[scanlation.PluginName] == nil { //TODO: refactor
		this.mapping[scanlation.PluginName] = make(map[JointScanlatorIds]scanlationIndex)
	}
	this.mapping[scanlation.PluginName][scanlation.Scanlators] = scanlationIndex(len(this.scanlations))
	this.scanlations = append(this.scanlations, scanlation)
	return false
}

func (this *Chapter) RemoveScanlation(index int) {
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

func (this *Chapter) RemoveScanlationsForPlugin(pluginName FetcherPluginName) {
	this.initialize()
	for _, realIndex := range this.mapping[pluginName] {
		this.scanlations = append(this.scanlations[:realIndex], this.scanlations[realIndex+1:]...)
	}
	delete(this.mapping, pluginName)
	deletionIndex, _ := qutils.IndexOf(this.usedPlugins, pluginName)
	this.usedPlugins = append(this.usedPlugins[:deletionIndex], this.usedPlugins[deletionIndex+1:]...)
}

func (this *Chapter) Scanlators() (ret []JointScanlatorIds) {
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

func (this *Chapter) SetParent(comic *Comic) {
	this.initialize()
	this.parent = comic
}

func (this *Chapter) SQLInsert(identity ChapterIdentity, stmts InsertionStmtGroup) (err error) {
	var newId int64
	result, err := stmts.chaptersInsertionStmt.Exec(this.sqlId, identity.n(), this.AlreadyRead)
	if err != nil {
		return err
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return err
	}
	this.sqlId = newId

	result, err = stmts.chaptersRelationStmt.Exec(this.parent.sqlId, this.sqlId)
	if err != nil {
		return err
	}

	for i := range this.scanlations {
		err = this.scanlations[i].SQLInsert(this.sqlId, stmts)
		if err != nil {
			return err
		}
	}

	return nil
}

func (this *Chapter) indexToPath(index int) (FetcherPluginName, JointScanlatorIds) {
	if this.parent == nil { //We have no parent, so we can't access priority lists for plugins and scanlators
		scanlation := this.scanlations[index]
		return scanlation.PluginName, scanlation.Scanlators
	}

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
			index -= jointsPerPlugin
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

func (this *Chapter) initialize() {
	if this.mapping == nil {
		this.mapping = make(map[FetcherPluginName]map[JointScanlatorIds]scanlationIndex)
	}
}

func (this *ChapterScanlation) SQLInsert(chapterId int64, stmts InsertionStmtGroup) (err error) {
	var newId int64
	result, err := stmts.scanlationInsertionStmt.Exec(this.sqlId, this.Title, this.Language, string(this.PluginName), this.URL)
	if err != nil {
		return err
	}
	newId, err = result.LastInsertId()
	if err != nil {
		return err
	}
	this.sqlId = newId

	result, err = stmts.scanlationRelationStmt.Exec(chapterId, this.sqlId)
	if err != nil {
		return err
	}

	for _, scanlator := range this.Scanlators.ToSlice() {
		result, err = stmts.scanlatorsRelationStmt.Exec(this.sqlId, scanlator)
		if err != nil {
			return err
		}
	}

	if this.plSQLIds == nil {
		this.plSQLIds = make([]int64, len(this.PageLinks))
	}
	for i, pageLink := range this.PageLinks {
		var pageLinkId int64 = this.plSQLIds[i] //WARNING: may go out of bounds (shouldn't ever; leaving it for the sake of experiment)
		result, err = stmts.pageLinksInsertionStmt.Exec(pageLink)
		if err != nil {
			return err
		}
		pageLinkId, err = result.LastInsertId()
		if err != nil {
			return err
		}
		this.plSQLIds[i] = pageLinkId
		stmts.pageLinksRelationStmt.Exec(this.sqlId, pageLinkId)
	}

	return nil
}
