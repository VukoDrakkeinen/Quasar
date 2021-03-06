package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/VukoDrakkeinen/Quasar/qutils"
)

const MATURE_GENRE_NAME = "Mature"

func MATURE_GENRE_ID() ComicGenreId {
	return ComicGenreId{1}
}

type ComicGenresDict struct {
	idAssigner
}

type ComicGenreId struct {
	ordinal Id
}

func NewComicGenresDict() ComicGenresDict {
	return ComicGenresDict{newIdAssigner(MATURE_GENRE_NAME)}
}

func (this *ComicGenresDict) AssignIds(genres []string) (ids []ComicGenreId, added []bool) {
	lids, added := this.idAssigner.assign(genres)
	for _, id := range lids {
		ids = append(ids, ComicGenreId{id})
	}
	return
}

func (this *ComicGenresDict) AssignIdsBytes(genres [][]byte) (ids []ComicGenreId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(genres))
}

func (this *ComicGenresDict) Id(genre string) ComicGenreId {
	return ComicGenreId{this.idAssigner.id(genre)}
}

func (this *ComicGenresDict) NameOf(id ComicGenreId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this ComicGenreId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), "TODO") //TODO
}

func (this ComicGenreId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}

func (this *ComicGenreId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New(fmt.Sprintf("%T.Scan: type assert failed (must be an int64, got %T!)", *this, src))
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
