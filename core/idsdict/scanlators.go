package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/qutils"
)

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
	return fmt.Sprintf("(%d)%s", int(this.ordinal), "TODO") //TODO
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

type JointScanlatorIds string //Can't have slices as keys in maps. Fortunately strings work, so we can pack data in them

func (this JointScanlatorIds) String() string {
	return fmt.Sprintf("%v", this.Slice())
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
