package idbase

import (
	"database/sql"
	"errors"
	"fmt"
	"quasar/qutils"
)

var Artists ArtistsDict

type ArtistsDict struct {
	idAssigner
}

type ArtistId struct {
	ordinal Id
}

func (this *ArtistsDict) AssignIds(artists []string) (ids []ArtistId, added []bool) {
	lids, added := this.idAssigner.assign(artists)
	for _, id := range lids {
		ids = append(ids, ArtistId{id})
	}
	return
}

func (this *ArtistsDict) AssignIdsBytes(artists [][]byte) (ids []ArtistId, added []bool) {
	return this.AssignIds(qutils.ByteSlicesToStrings(artists))
}

func (this *ArtistsDict) Id(artist string) ArtistId {
	return ArtistId{this.idAssigner.id(artist)}
}

func (this *ArtistsDict) NameOf(id ArtistId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this ArtistId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Artists.NameOf(this))
}

func (this ArtistId) ExecuteInsertionStmt(stmt *sql.Stmt, IinfoId ...interface{}) (err error) {
	if len(IinfoId) != 1 {
		return errors.New("ArtistId.ExecuteDBStatement: invalid number of parameters!")
	}
	for _, infoId := range IinfoId {
		_, err = stmt.Exec(infoId, this.ordinal+1) //RDBMSes start counting at 1, not 0
	}
	return
}

func (this *ArtistId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("ArtistId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}
