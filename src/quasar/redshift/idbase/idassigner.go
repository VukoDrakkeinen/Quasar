package idbase

import (
	"bytes"
	. "database/sql" //WORKAROUND: imports analyzer complains about unused one
	"encoding/json"
	"quasar/redshift/qdb"
	"strconv"
	"strings"
	"sync"
)

type IdAssigner struct {
	names   []string
	mapping map[string]Id
	lock    sync.Mutex

	preLoadedNames []string
}
type Id int

type assignerJsonProxy struct {
	Names []string
}

func (this *IdAssigner) reset() {
	this.names = make([]string, 0, 10)
	this.mapping = make(map[string]Id)
	this.assign([]string{"Unknown"})
	this.assign(this.preLoadedNames)
}

func (this *IdAssigner) initialize() *IdAssigner {
	if this.mapping == nil {
		this.reset()
		//this.load()
	}
	return this
}

func (this *IdAssigner) toJSONProxy() *assignerJsonProxy {
	return &assignerJsonProxy{Names: this.names}
}

func (this *IdAssigner) loadData(b []byte, fieldName string) {
	var proxy assignerJsonProxy
	err := json.Unmarshal(bytes.Replace(b, []byte(strconv.Quote(fieldName)), []byte(`"Names"`), 1), &proxy)
	if err != nil {
		//TODO: log error
	} else {
		this.reset()
		this.assign(proxy.Names)
	}
}

func (this *IdAssigner) getSaveData(fieldName string) []byte {
	this.initialize()
	this.lock.Lock()
	defer this.lock.Unlock()
	jsonData, _ := json.Marshal(this.toJSONProxy())
	var buf bytes.Buffer
	json.Indent(&buf, bytes.Replace(jsonData, []byte(`"Names"`), []byte(strconv.Quote(fieldName)), 1), "", "\t")
	return buf.Bytes()
}

func (this *IdAssigner) saveToDB(table string) {
	db := qdb.DB() //TODO: error out on nil
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ids.` + table + ` (names TEXT UNIQUE NOT NULL);`)
	if err != nil {
		//TODO: log error
		return
	}

	transaction, err := db.Begin()
	if err != nil {
		//TODO: log error
	}
	statement, err := transaction.Prepare(`INSERT OR IGNORE INTO ids.` + table + `(names) values(?);`)
	if err != nil {
		//TODO: log error
	}
	defer statement.Close()
	for name := range this.names {
		_, err = statement.Exec(name)
		if err != nil {
			//TODO: log error
		}
	}
	transaction.Commit()
}

func (this *IdAssigner) loadFromDB(table string) {
	db := qdb.DB() //TODO: nil!
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS ids.` + table + `(names TEXT UNIQUE NOT NULL);`)
	if err != nil { //TODO: copy-pasted code
		//TODO: log error
		return
	}

	statement, err := db.Prepare(`SELECT names FROM ids.` + table)
	if err != nil {
		//TODO: log error
	}
	defer statement.Close()
	rows, err := statement.Query()
	if err != nil {
		//TODO: log error
	}
	defer rows.Close()
	names := make([]string, 0, 8)
	for rows.Next() {
		var name string
		rows.Scan(&name)
		names = append(names, name)
	}
	this.reset()
	this.assign(names)
}

func (this *IdAssigner) assign(names []string) (ids []Id, added []bool) {
	this.initialize()
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, name := range names {
		name = strings.ToLower(name)
		_, exists := this.mapping[name]
		if !exists {
			this.names = append(this.names, name)
			this.mapping[name] = Id(len(this.names) - 1)
		}
		added = append(added, !exists)
		ids = append(ids, this.mapping[name])
	}
	return
}

func (this *IdAssigner) id(name string) Id {
	this.initialize()
	if id, exists := this.mapping[strings.ToLower(name)]; exists {
		return id
	}
	return Id(0) //Unknown
}

func (this *IdAssigner) nameOf(id Id) string {
	this.initialize()
	return strings.Replace(strings.Title(this.names[id]), "/S", "/s", -1) //FIXME: strings.Title() has a bug where it will occasionally capitalize wrong letters
}
