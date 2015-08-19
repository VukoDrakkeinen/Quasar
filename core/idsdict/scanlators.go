package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"sort"
	"unsafe"
)

var Scanlators = NewScanlatorsDict()

type ScanlatorsDict struct {
	idAssigner
}

type ScanlatorId struct {
	ordinal Id
}

func NewScanlatorsDict() ScanlatorsDict {
	return ScanlatorsDict{newIdAssigner()}
}

func (this *ScanlatorsDict) AssignIds(scanlators []string) (ids []ScanlatorId, added []bool) {
	lids, added := this.idAssigner.assign(scanlators)
	for _, id := range lids {
		ids = append(ids, ScanlatorId{id})
	}
	return
}

func (this *ScanlatorsDict) AssignIdsBytes(scanlators [][]byte) (ids []ScanlatorId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(scanlators))
}

func (this *ScanlatorsDict) Id(scanlator string) ScanlatorId {
	return ScanlatorId{this.idAssigner.id(scanlator)}
}

func (this *ScanlatorsDict) NameOf(id ScanlatorId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this ScanlatorId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Scanlators.NameOf(this))
}

func (this *ScanlatorId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be an int64, got %T!)", *this, src))
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}

func (this ScanlatorId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}

type JointScanlatorIds struct {
	data  string //Can't have slices as keys in maps. Fortunately strings work, so we can pack data in them
	count int
}

type ScanlatorSlice []ScanlatorId

func (slice ScanlatorSlice) Len() int {
	return len(slice)
}

func (slice ScanlatorSlice) Less(i, j int) bool {
	return slice[i].ordinal < slice[j].ordinal
}

func (slice ScanlatorSlice) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func JoinScanlators(ids []ScanlatorId) JointScanlatorIds {
	sort.Sort(ScanlatorSlice(ids))
	runes := make([]rune, 0, len(ids))
	for _, id := range ids {
		runes = append(runes, rune(id.ordinal)) //possibly narrowing from 64 to 32 bits, shouldn't matter in the long run
	}
	return JointScanlatorIds{data: *(*string)(unsafe.Pointer(&runes)), count: len(ids)}

}

func (this *JointScanlatorIds) ToSlice() []ScanlatorId {
	ids := make([]ScanlatorId, 0, this.count)
	for _, drune := range *(*[]rune)(unsafe.Pointer(&this.data)) {
		ids = append(ids, ScanlatorId{Id(drune)})
	}
	return ids
}

func (this JointScanlatorIds) String() string {
	return fmt.Sprintf("%v", this.ToSlice())
}

/*
const ratio = int(unsafe.Sizeof(Id(0)) / unsafe.Sizeof(' '))	//architecture-independent implementation

func JoinScanlators_(ids []ScanlatorId) JointScanlatorIds {
	sort.Sort(ScanlatorSlice(ids))
	runes := make([]rune, 0, len(ids)*ratio)
	switch ratio {
	case 1:
	for _, id := range ids {
		runes = append(runes, rune(id.ordinal))
	}
	case 2:
	for _, id := range ids {
		runes = append(runes, rune(id.ordinal))
		runes = append(runes, rune(id.ordinal>>32))
	}
	}
	return JointScanlatorIds{data: *(*string)(unsafe.Pointer(&runes)), count: len(ids)}
}

func (this *JointScanlatorIds) ToSlice_() []ScanlatorId {
	ids := make([]ScanlatorId, 0, this.count)
	switch ratio {
	case 1:
	for _, drune := range *(*[]rune)(unsafe.Pointer(&this.data)) {
		ids = append(ids, ScanlatorId{Id(drune)})
	}
	case 2:
		var rune0 rune
	for i, drune := range *(*[]rune)(unsafe.Pointer(&this.data)) {
		switch i % 2 {
		case 0:
			rune0 = drune
		case 1:
			ids = append(ids, ScanlatorId{(Id(drune) << 32) | Id(rune0)})
			rune0 = 0
		}
	}
	}
	return ids
}//*/
