package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/qutils"
)

type AuthorsDict struct {
	idAssigner
}

type AuthorId struct {
	ordinal Id
}

func NewAuthorDict() AuthorsDict {
	return AuthorsDict{newIdAssigner()}
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
	return fmt.Sprintf("(%d)%s", int(this.ordinal), "TODO") //TODO
}

func (this AuthorId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}

func (this *AuthorId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be an int64, got %T!)", *this, src))
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
