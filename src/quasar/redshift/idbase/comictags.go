package idbase

import (
	"database/sql"
	"errors"
	"fmt"
	"quasar/qutils"
)

var ComicTags ComicTagsDict

type ComicTagsDict struct {
	idAssigner
}

type ComicTagId struct {
	ordinal Id
}

func (this *ComicTagsDict) AssignIds(tags []string) (ids []ComicTagId, added []bool) {
	lids, added := this.idAssigner.assign(tags)
	for _, id := range lids {
		ids = append(ids, ComicTagId{id})
	}
	return
}

func (this *ComicTagsDict) AssignIdsBytes(tags [][]byte) (ids []ComicTagId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(tags))
}

func (this *ComicTagsDict) Id(tag string) ComicTagId {
	return ComicTagId{this.idAssigner.id(tag)}
}

func (this *ComicTagsDict) NameOf(id ComicTagId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this ComicTagId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), ComicTags.NameOf(this))
}

func (this ComicTagId) ExecuteInsertionStmt(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		panic("ComicTagId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1 not 0
	}
	return
}

func (this *ComicTagId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ComicTagId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
