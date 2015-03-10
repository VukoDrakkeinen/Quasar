package idbase

import (
	"database/sql"
	"fmt"
	"quasar/qutils"
)

var ComicTags ComicTagsDict

type ComicTagsDict struct {
	IdAssigner
}

type ComicTagId struct {
	ordinal Id
}

func (this *ComicTagsDict) AssignIds(tags []string) (ids []ComicTagId, added []bool) {
	lids, added := this.IdAssigner.assign(tags)
	for _, id := range lids {
		ids = append(ids, ComicTagId{id})
	}
	return
}

func (this *ComicTagsDict) AssignIdsBytes(tags [][]byte) (ids []ComicTagId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(tags))
}

func (this *ComicTagsDict) Id(tag string) ComicTagId {
	return ComicTagId{this.IdAssigner.id(tag)}
}

func (this *ComicTagsDict) NameOf(id ComicTagId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this ComicTagId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), ComicTags.NameOf(this))
}

func (this ComicTagId) ExecuteDBStatement(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		panic("ComicTagId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1 not 0
	}
	return
}
