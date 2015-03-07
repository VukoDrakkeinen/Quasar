package idbase

import (
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
	IdAssigner
}

type ComicGenreId struct {
	ordinal Id
}

func (this *ComicGenresDict) AssignIds(genres []string) (ids []ComicGenreId, added []bool) {
	lids, added := this.IdAssigner.assign(genres)
	for _, id := range lids {
		ids = append(ids, ComicGenreId{id})
	}
	return
}

func (this *ComicGenresDict) AssignIdsBytes(genres [][]byte) (ids []ComicGenreId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(genres))
}

func (this *ComicGenresDict) Id(genre string) ComicGenreId {
	return ComicGenreId{this.IdAssigner.id(genre)}
}

func (this *ComicGenresDict) NameOf(id ComicGenreId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this ComicGenreId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), ComicGenres.NameOf(this))
}

func (this *ComicGenresDict) Save() {
	this.IdAssigner.saveToDB("genres")
}

func (this *ComicGenresDict) Load() {
	this.IdAssigner.loadFromDB("genres")
}
