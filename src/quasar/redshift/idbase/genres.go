package idbase

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"quasar/qutils"
)

var ComicGenres ComicGenresDict

const MATURE_GENRE_NAME = "Mature"

func init() {
	ComicGenres.AssignIds([]string{MATURE_GENRE_NAME})
}

func MATURE_GENRE() ComicGenreId {
	return ComicGenres.Id(MATURE_GENRE_NAME)
}

type ComicGenresDict struct {
	idAssigner
}

type ComicGenreId struct {
	ordinal Id
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
	return fmt.Sprintf("(%d)%s", int(this.ordinal), ComicGenres.NameOf(this))
}

func (this ComicGenreId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}

func (this *ComicGenreId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ComicGenreId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
