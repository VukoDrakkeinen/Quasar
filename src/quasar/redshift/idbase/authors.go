package idbase

import (
	"database/sql"
	"errors"
	"fmt"
	"quasar/qutils"
)

var Authors AuthorsDict

type AuthorsDict struct {
	idAssigner
}

type AuthorId struct {
	ordinal Id
}

func (this *AuthorsDict) AssignIds(authors []string) (ids []AuthorId, added []bool) {
	lids, added := this.idAssigner.assign(authors)
	for _, id := range lids {
		ids = append(ids, AuthorId{id})
	}
	return
}

func (this *AuthorsDict) AssignIdsBytes(authors [][]byte) (ids []AuthorId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(authors))
}

func (this *AuthorsDict) Id(author string) AuthorId {
	return AuthorId{this.idAssigner.id(author)}
}

func (this *AuthorsDict) NameOf(id AuthorId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this AuthorId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Authors.NameOf(this))
}

func (this AuthorId) ExecuteInsertionStmt(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		return errors.New("AuthorId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1, not 0
	}
	return
}

func (this *AuthorId) Scan(src interface{}) error {
	n, ok := src.(int64) //TODO?: check if scanned id is assigned
	if !ok || src == nil {
		return errors.New("AuthorId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
