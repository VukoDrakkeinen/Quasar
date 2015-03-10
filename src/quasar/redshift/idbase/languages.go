package idbase

import (
	"database/sql"
	"fmt"
)

const ENGLISH_LANG_NAME = "English"

var LangDict LanguageDict

func init() {
	LangDict.AssignIds([]string{ENGLISH_LANG_NAME})
	LangDict.preLoadedNames = []string{ENGLISH_LANG_NAME}
}

func ENGLISH_LANG() LangId {
	return LangDict.Id(ENGLISH_LANG_NAME)
}

type LanguageDict struct {
	IdAssigner
}

type LangId struct {
	ordinal Id
}

func (this *LanguageDict) AssignIds(langs []string) (ids []LangId, added []bool) {
	lids, added := this.IdAssigner.assign(langs)
	for _, id := range lids {
		ids = append(ids, LangId{id})
	}
	return
}

func (this *LanguageDict) Id(lang string) LangId {
	return LangId{this.IdAssigner.id(lang)}
}

func (this *LanguageDict) NameOf(id LangId) string {
	return this.IdAssigner.nameOf(id.ordinal)
}

func (this LangId) String() string {
	return fmt.Sprintf("(%d)%s", int(this.ordinal), LangDict.NameOf(this))
}

func (this LangId) ExecuteDBStatement(stmt *sql.Stmt, scanlationData ...interface{}) (err error) {
	if len(scanlationData) != 3 {
		panic("LangId.ExecuteDBStatement: invalid number of parameters!")
	}
	title := scanlationData[0].(string)
	pluginName := scanlationData[1].(string)
	url := scanlationData[2].(string)
	_, err = stmt.Exec(title, this.ordinal+1, pluginName, url) //RDBMSes start counting at 1 not 0
	return
}

/*
const languageBaseFile = "langs.json"

func (this *LanguageDict) Save() {
	WriteConfig(languageBaseFile, this.IdAssigner.getSaveData("languages"))
}

func (this *LanguageDict) Load() {
	data, err := ReadConfig(languageBaseFile)
	if err != nil {
		//tTODO: log error
	} else {
		this.IdAssigner.loadData(data, "languages")
	}
}
*/
