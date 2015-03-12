package idbase

import (
	"database/sql"
	"errors"
	"fmt"
	"quasar/qutils"
	"strconv"
	"strings"
)

var Scanlators ScanlatorsDict

type ScanlatorsDict struct {
	idAssigner
}

type ScanlatorId struct {
	ordinal Id
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

func (this ScanlatorId) ExecuteInsertionStmt(stmt *sql.Stmt, IscanlationId ...interface{}) (err error) {
	if len(IscanlationId) != 1 {
		panic("ScanlatorId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, scanlationId := range IscanlationId {
		_, err = stmt.Exec(scanlationId, this.ordinal+1) //RDBMSes start counting at 1 not 0
	}
	return
}

func (this *ScanlatorId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ScanlatorId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
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
	ids := this.ToSlice() /*
		paramsConf := strings.Repeat("%v ", len(ids))
		paramsConf = paramsConf[:len(paramsConf)-1] //remove trailing space
		paramsConf = fmt.Sprintf("[%s]", paramsConf)
		return fmt.Sprintf(paramsConf, ids)	//*/
	return fmt.Sprintf("%v", ids)
}
