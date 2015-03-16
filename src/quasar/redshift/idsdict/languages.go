package idsdict

import (
	"database/sql/driver"
	"errors"
	"fmt"
)

const ENGLISH_LANG_NAME = "English"

var Langs = NewLangDict()

func ENGLISH_LANG() LangId {
	return Langs.Id(ENGLISH_LANG_NAME)
}

type LangsDict struct {
	idAssigner
}

type LangId struct {
	ordinal Id
}

func NewLangDict() LangsDict {
	return LangsDict{newIdAssigner(ENGLISH_LANG_NAME)}
}

func (this *LangsDict) AssignIds(langs []string) (ids []LangId, added []bool) {
	lids, added := this.idAssigner.assign(langs)
	for _, id := range lids {
		ids = append(ids, LangId{id})
	}
	return
}

func (this *LangsDict) Id(lang string) LangId {
	return LangId{this.idAssigner.id(lang)}
}

func (this *LangsDict) NameOf(id LangId) string {
	return this.idAssigner.nameOf(id.ordinal)
}

func (this LangId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), Langs.NameOf(this))
}

func (this *LangId) Scan(src interface{}) error {
	n, ok := src.(int64)
	if !ok || src == nil {
		return errors.New("LangId.Scan: type assert failed (must be an int64!)")
	}
	this.ordinal = Id(n - 1) //RDBMSes start counting at 1, not 0
	return nil
}

func (this LangId) Value() (driver.Value, error) {
	return int64(this.ordinal + 1), nil //RDBMSes start counting at 1, not 0
}
