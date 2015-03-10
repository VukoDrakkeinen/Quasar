package idbase

import (
	"database/sql"
	"fmt"
	"quasar/qutils"
)

var Authors AuthorsDict

type AuthorsDict struct {
	IdAssigner
}

type AuthorId struct {
	ordinal Id
}

func (this *AuthorsDict) AssignIds(authors []string) (ids []AuthorId, added []bool) {
	lids, added := this.IdAssigner.assign(authors)
	for _, id := range lids {
		ids = append(ids, AuthorId{id})
	}
	return
}

func (this *AuthorsDict) AssignIdsBytes(authors [][]byte) (ids []AuthorId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(authors))
}

func (this *AuthorsDict) Id(author string) AuthorId {
	return AuthorId{this.IdAssigner.id(author)}
}

func (this *AuthorsDict) NameOf(id AuthorId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this AuthorId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Authors.NameOf(this))
}

func (this AuthorId) ExecuteDBStatement(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		panic("AuthorId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1 not 0
	}
	return
}
