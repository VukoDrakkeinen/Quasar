package idsdict

import (
	"database/sql"
	"strings"
	"sync"
)

type idAssigner struct {
	names         []string
	mapping       map[string]Id
	constantNames []string
	lock          sync.RWMutex

	replacer *strings.Replacer //FIXME: see bug #233830
}
type Id int

func newIdAssigner(constantNames ...string) idAssigner {
	ret := idAssigner{
		replacer:      strings.NewReplacer("/S", "/s", "'S", "'s"), //FIXME: see bug #233830
		constantNames: constantNames,
	}
	ret.reset()
	return ret
}

func (this *idAssigner) ExecuteInsertionStmt(stmt *sql.Stmt, unused ...interface{}) error {
	_ = unused
	this.lock.RLock()
	defer this.lock.RUnlock()
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
	this.lock.Lock()
	this.names = make([]string, 0, 10)
	this.mapping = make(map[string]Id)
	this.lock.Unlock()
	this.assign(append([]string{"Unknown"}, this.constantNames...))
}

func (this *idAssigner) assign(names []string) (ids []Id, added []bool) {
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
	this.lock.RLock()
	defer this.lock.RUnlock()
	if id, exists := this.mapping[strings.ToLower(name)]; exists {
		return id
	}
	return Id(0) //Unknown
}

func (this *idAssigner) nameOf(id Id) string {
	this.lock.RLock()
	defer this.lock.RUnlock()
	if int(id) >= len(this.names) {
		id = Id(0)
	}
	//FIXME: bug #233830: strings.Title() has a bug where it will occasionally capitalize wrong letters
	return this.replacer.Replace(strings.Title(this.names[id]))
}
