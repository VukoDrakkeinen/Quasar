package idbase

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"sync"
)

type idAssigner struct {
	names   []string
	mapping map[string]Id
	lock    sync.Mutex

	preLoadedNames []string

	replacer *strings.Replacer //FIXME: see bug #233830
}
type Id int

type assignerJsonProxy struct {
	Names []string
}

func (this *idAssigner) ExecuteInsertionStmt(stmt *sql.Stmt, unused ...interface{}) error {
	_ = unused
	for _, name := range this.names {
		_, err := stmt.Exec(name)
		if err != nil {
			return err
		}
	}
	return nil
}

func (this *idAssigner) ExecuteQueryStmt(stmt *sql.Stmt, unused ...interface{}) error {
	_ = unused
	nameRows, err := stmt.Query()
	if err != nil {
		return err
	}
	var names []string
	for nameRows.Next() {
		var name string
		nameRows.Scan(&name)
		names = append(names, name)
	}
	this.reset()
	this.assign(names)
	return nil
}

func (this *idAssigner) reset() {
	this.names = make([]string, 0, 10)
	this.mapping = make(map[string]Id)
	this.assign([]string{"Unknown"})
	this.assign(this.preLoadedNames)
}

func (this *idAssigner) initialize() *idAssigner {
	if this.mapping == nil {
		this.reset()
		this.replacer = strings.NewReplacer("/S", "/s", "'S", "'s") //FIXME: see bug #233830
		//this.load()
	}
	return this
}

func (this *idAssigner) toJSONProxy() *assignerJsonProxy {
	return &assignerJsonProxy{Names: this.names}
}

func (this *idAssigner) loadData(b []byte, fieldName string) {
	var proxy assignerJsonProxy
	err := json.Unmarshal(bytes.Replace(b, []byte(strconv.Quote(fieldName)), []byte(`"Names"`), 1), &proxy)
	if err != nil {
		//TODO: log error
	} else {
		this.reset()
		this.assign(proxy.Names)
	}
}

func (this *idAssigner) getSaveData(fieldName string) []byte {
	this.initialize()
	this.lock.Lock()
	defer this.lock.Unlock()
	jsonData, _ := json.Marshal(this.toJSONProxy())
	var buf bytes.Buffer
	json.Indent(&buf, bytes.Replace(jsonData, []byte(`"Names"`), []byte(strconv.Quote(fieldName)), 1), "", "\t")
	return buf.Bytes()
}

func (this *idAssigner) assign(names []string) (ids []Id, added []bool) {
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

func (this *idAssigner) id(name string) Id {
	this.initialize()
	if id, exists := this.mapping[strings.ToLower(name)]; exists {
		return id
	}
	return Id(0) //Unknown
}

func (this *idAssigner) nameOf(id Id) string {
	this.initialize()
	//FIXME: bug #233830: strings.Title() has a bug where it will occasionally capitalize wrong letters
	return this.replacer.Replace(strings.Title(this.names[id]))
}
