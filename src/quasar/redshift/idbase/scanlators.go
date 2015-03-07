package idbase

import (
	"fmt"
	"quasar/qutils"
	"strconv"
	"strings"
)

var Scanlators ScanlatorsDict

type ScanlatorsDict struct {
	IdAssigner
}

type ScanlatorId struct {
	ordinal Id
}

func (this *ScanlatorsDict) AssignIds(scanlators []string) (ids []ScanlatorId, added []bool) {
	lids, added := this.IdAssigner.assign(scanlators)
	for _, id := range lids {
		ids = append(ids, ScanlatorId{id})
	}
	return
}

func (this *ScanlatorsDict) AssignIdsBytes(scanlators [][]byte) (ids []ScanlatorId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(scanlators))
}

func (this *ScanlatorsDict) Id(scanlator string) ScanlatorId {
	return ScanlatorId{this.IdAssigner.id(scanlator)}
}

func (this *ScanlatorsDict) NameOf(id ScanlatorId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this ScanlatorId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Scanlators.NameOf(this))
}

func (this *ScanlatorsDict) Save() {
	this.IdAssigner.saveToDB("scanlators")
}

func (this *ScanlatorsDict) Load() {
	this.IdAssigner.loadFromDB("scanlators")
}

//////////////

type JointScanlatorIds struct { //Can't have slices as keys in maps? Here's a dirty ha-, I mean, a workaround for you!
	data  string
	count int
}

func JoinScanlators(ids []ScanlatorId) JointScanlatorIds {
	stringNums := make([]string, 0, len(ids))
	for _, id := range ids {
		stringNums = append(stringNums, strconv.FormatInt(int64(id.ordinal), 10))
	}
	return JointScanlatorIds{data: strings.Join(stringNums, "&"), count: len(ids)}
}

func (this *JointScanlatorIds) ToSlice() []ScanlatorId {
	ids := make([]ScanlatorId, 0, this.count)
	for _, stringNum := range strings.Split(this.data, "&") {
		num, _ := strconv.ParseInt(stringNum, 10, 32)
		ids = append(ids, ScanlatorId{Id(num)})
	}
	return ids
}

func (this JointScanlatorIds) String() string {
	ids := this.ToSlice()
	paramsConf := strings.Repeat("%v ", len(ids))
	paramsConf = paramsConf[:len(paramsConf)-1] //remove trailing space
	paramsConf = fmt.Sprintf("[%s]", paramsConf)
	return fmt.Sprintf(paramsConf, ids)
}
