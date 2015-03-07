package redshift

import (
	"fmt"
	"quasar/qutils"
	. "quasar/redshift/idbase"
)

type Chapter struct {
	parent      *Comic
	data        map[FetcherPluginName]ChapterDataHolder
	AlreadyRead bool
}

func (this *Chapter) initialize() {
	if this.data == nil {
		this.data = make(map[FetcherPluginName]ChapterDataHolder)
	}
}

func (this *Chapter) SetParent(comic *Comic) {
	this.initialize()
	this.parent = comic
}

func (this *Chapter) Scanlators() []ScanlatorId {
	this.initialize()
	ret := make([]ScanlatorId, 0)
	for _, holder := range this.data {
		for _, joint := range holder.order {
			ret = append(ret, joint.ToSlice()...)
		}
		//ret = append(ret, holder.order...)
	}
	return ret
}

func (this *Chapter) DataCount() int {
	this.initialize()
	count := 0
	for _, holder := range this.data {
		count += len(holder.dataSlice)
	}
	return count
}

func (this *Chapter) AltersCount() map[FetcherPluginName]int {
	this.initialize()
	ret := make(map[FetcherPluginName]int)
	for pluginName, holder := range this.data {
		ret[pluginName] = len(holder.dataSlice)
	}
	return ret
}

func (this *Chapter) Data(index int) ChapterData {
	this.initialize()
	if this.parent != nil { //we have a parent Comic, so we can access priority lists for plugins and scanlators
		scIndex := index
		plIndex := 0
		sources := this.parent.Sources()
		for dsLen := len(this.data[sources[plIndex].PluginName].dataSlice); index >= dsLen; {
			//previous plugin's Holder doesn't contain that many ChapterDatas
			plIndex++        //so go to the next one
			scIndex -= dsLen //and skip all the previous ChapterDatas
		}
		holder := this.data[sources[plIndex].PluginName]

		//reorder scanlators according to priority (parent.priority first, then the rest)
		parentScPrio := this.parent.ScanlatorsPriority()
		scanlatorPriority := make([]JointScanlatorIds, 0, len(parentScPrio)+len(holder.order))
		alreadyAddedAndSome := make(map[JointScanlatorIds]struct{})
		for _, scanlator := range parentScPrio {
			alreadyAddedAndSome[scanlator] = struct{}{}
			if _, exists := holder.indexByScanlators[scanlator]; exists { //add only shared
				scanlatorPriority = append(scanlatorPriority, scanlator)
			}
		}
		for _, scanlator := range holder.order {
			if _, exists := alreadyAddedAndSome[scanlator]; !exists { //add the rest, skip already added
				scanlatorPriority = append(scanlatorPriority, scanlator)
			}
		}
		return holder.dataSlice[holder.indexByScanlators[scanlatorPriority[scIndex]]]
	} else { //Unknown preferred order of plugins, use random! D:
		var dset ChapterDataHolder
		for key := range this.data {
			dset = this.data[key]
			break
		}
		return dset.dataSlice[dset.indexByScanlators[dset.order[index]]]
	}
}

func (this *Chapter) DataForPlugin(pluginName FetcherPluginName, index int) ChapterData {
	this.initialize()
	fmt.Println(this.data)
	fmt.Println(this.data[pluginName])
	return this.data[pluginName].dataSlice[index]
}

func (this *Chapter) SetData(pluginName FetcherPluginName, data ChapterData) {
	this.initialize()
	holder := this.data[pluginName]
	if holder.indexByScanlators == nil {
		holder.indexByScanlators = make(map[JointScanlatorIds]DataIndex)
	}
	if index, exists := holder.indexByScanlators[data.Scanlators]; exists {
		holder.dataSlice[index] = data
	} else {
		holder.dataSlice = append(holder.dataSlice, data)
		holder.order = append(holder.order, data.Scanlators)
		holder.indexByScanlators[data.Scanlators] = DataIndex(len(holder.dataSlice) - 1)
	}
	this.data[pluginName] = holder
}

func (this *Chapter) Merge(another *Chapter) {
	this.initialize()
	this.AlreadyRead = another.AlreadyRead || this.AlreadyRead
	for pluginName := range another.data {
		if _, exists := this.data[pluginName]; !exists {
			this.data[pluginName] = another.data[pluginName]
		} else {
			holder0 := this.data[pluginName]
			holder1 := another.data[pluginName]
			for _, scanlator := range holder1.order { //Merge holders
				if _, exists := holder0.indexByScanlators[scanlator]; !exists {
					newIndex := holder1.indexByScanlators[scanlator]
					holder0.indexByScanlators[scanlator] = newIndex
					holder0.order = append(holder0.order, scanlator)
					holder0.dataSlice = append(holder0.dataSlice, holder1.dataSlice[newIndex])
				}
			}
			this.data[pluginName] = holder0
		}
	}
}

type DataIndex int
type ChapterDataHolder struct {
	dataSlice         []ChapterData
	order             []JointScanlatorIds
	indexByScanlators map[JointScanlatorIds]DataIndex //also scanlatorSet
}

type ChapterData struct {
	Scanlators JointScanlatorIds //TODO: refactor, shit's confusing (and scanlators list doubles in ChpDatas)
	Title      string
	Language   LangId
	URL        string
	PageLinks  []string
}

const (
	LQ_Modifier byte = 10 * iota
	MQ_Modifier
	HQ_Modifier
)

type ChapterIdentity struct {
	Volume   byte
	MajorNum uint16
	MinorNum byte
	Letter   byte
	Version  byte
}

func (this ChapterIdentity) Equals(another ChapterIdentity) bool {
	return this.Volume != another.Volume ||
		this.MajorNum != another.MajorNum ||
		this.MinorNum != another.MinorNum ||
		this.Letter != another.Letter ||
		this.Version != another.Version
}

func (this ChapterIdentity) Less(another ChapterIdentity) bool {
	return this.Volume < another.Volume ||
		(this.Volume == another.Volume && this.MajorNum < another.MajorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum < another.MinorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter < another.Letter) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter == another.Letter && this.Version < this.Version)
}

func (this ChapterIdentity) More(another ChapterIdentity) bool {
	return this.Volume > another.Volume ||
		(this.Volume == another.Volume && this.MajorNum > another.MajorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum > another.MinorNum) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter > another.Letter) ||
		(this.Volume == another.Volume && this.MajorNum == another.MajorNum && this.MinorNum == another.MinorNum && this.Letter == another.Letter && this.Version > this.Version)
}

func (this ChapterIdentity) LessEq(another ChapterIdentity) bool {
	return !this.More(another)
}

func (this ChapterIdentity) MoreEq(another ChapterIdentity) bool {
	return !this.Less(another)
}

func (this ChapterIdentity) n() int64 {
	return int64(this.Volume)<<40 | int64(this.MajorNum)<<24 | int64(this.MinorNum)<<16 | int64(this.Letter)<<8 | int64(this.Version)
}

type ChapterIdentitiesSlice []ChapterIdentity

func (this ChapterIdentitiesSlice) Len() int {
	return len(this)
}

func (this ChapterIdentitiesSlice) Less(i, j int) bool {
	return this[i].Less(this[j])
}

func (this ChapterIdentitiesSlice) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this ChapterIdentitiesSlice) overflowingSearch(ci ChapterIdentity) (index int64) {
	low := int64(0)
	mid := int64(0)
	high := int64(0)
	if this.Len() > 0 {
		high = int64(this.Len() - 1)
	}

	if high == 0 || ci.More(this[high]) {
		return int64(this.Len())
	}
	for this[low].LessEq(ci) && this[high].MoreEq(ci) {
		mid = low + ((ci.n()-this[low].n())*(high-low))/(this[high].n()-this[low].n())
		if this[mid].Less(ci) {
			low = mid + 1
		} else if this[mid].More(ci) {
			high = mid - 1
		} else {
			return mid
		}
	}
	if this[low].Equals(ci) {
		return low
	} else if ci.More(this[low]) {
		return mid
	} else {
		return low
	}
}

func (this ChapterIdentitiesSlice) Search(ci ChapterIdentity) (index int64) {
	low := int64(0)
	mid := int64(0)
	high := int64(0)
	if this.Len() > 0 {
		high = int64(this.Len() - 1)
	}
	if high == 0 || ci.n() > this[high].n() {
		return int64(this.Len())
	}
	for this[low].n() <= ci.n() && this[high].n() >= ci.n() {
		diff0 := ci.n() - this[low].n()
		diff1 := high - low
		diff2 := this[high].n() - this[low].n()
		mid = low + qutils.MultiplyThenDivide(diff0, diff1, diff2)
		if this[mid].n() < ci.n() {
			low = mid + 1
		} else if this[mid].n() > ci.n() {
			high = mid - 1
		} else {
			return mid
		}
	}

	if ci.n() > this[low].n() {
		return mid
	} else {
		return low
	}
}

func (this ChapterIdentitiesSlice) Insert(at int, ci ChapterIdentity) ChapterIdentitiesSlice {
	this = append(this, ChapterIdentity{})
	copy(this[at+1:], this[at:])
	this[at] = ci
	return this
}

func (this ChapterIdentitiesSlice) InsertMultiple(at int, cis []ChapterIdentity) ChapterIdentitiesSlice {
	newSize := len(this) + len(cis)
	if cap(this) < newSize {
		newSlice := make([]ChapterIdentity, newSize, qutils.GrownCap(newSize))
		copy(newSlice, this[:at])
		copy(newSlice[at:], cis)
		copy(newSlice[at+len(cis):], this[at:])
		return newSlice
	} else {
		this = this[:newSize]
		copy(this[at+len(cis):], this[at:])
		copy(this[at:], cis)
		return this
	}
}
