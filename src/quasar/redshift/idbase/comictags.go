package idbase

import (
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

func (this *ComicTagsDict) Save() {
	this.IdAssigner.saveToDB("categories")
}

func (this *ComicTagsDict) Load() {
	this.IdAssigner.loadFromDB("categories")
}
