package idbase

import (
	"fmt"
	"quasar/qutils"
)

var Artists ArtistsDict

type ArtistsDict struct {
	IdAssigner
}

type ArtistId struct {
	ordinal Id
}

func (this *ArtistsDict) AssignIds(artists []string) (ids []ArtistId, added []bool) {
	lids, added := this.IdAssigner.assign(artists)
	for _, id := range lids {
		ids = append(ids, ArtistId{id})
	}
	return
}

func (this *ArtistsDict) AssignIdsBytes(artists [][]byte) (ids []ArtistId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(artists))
}

func (this *ArtistsDict) Id(artist string) ArtistId {
	return ArtistId{this.IdAssigner.id(artist)}
}

func (this *ArtistsDict) NameOf(id ArtistId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this ArtistId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Artists.NameOf(this))
}

func (this *ArtistsDict) Save() {
	this.IdAssigner.saveToDB("artists")
}

func (this *ArtistsDict) Load() {
	this.IdAssigner.loadFromDB("artists")
}
