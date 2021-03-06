package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/qutils"
)

type ComicTagsDict struct {
	idAssigner
}

type ComicTagId struct {
	ordinal Id
}

func NewComicTagsDict() ComicTagsDict {
	return ComicTagsDict{newIdAssigner()}
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
	return fmt.Sprintf("(%d)%s", int(this.ordinal), "TODO") //TODO
}

func (this ComicTagId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}

func (this *ComicTagId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be an int64, got %T!)", *this, src))
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
