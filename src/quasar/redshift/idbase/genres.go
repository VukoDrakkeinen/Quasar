package idbase

import (
	"database/sql"
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

func (this ComicGenreId) ExecuteInsertionStmt(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		panic("ComicGenreId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1 not 0
	}
	return
}

func (this *ComicGenreId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ComicGenreId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
