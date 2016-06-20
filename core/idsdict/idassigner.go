package idsdict

import (
	"database/sql"
	"github.com/VukoDrakkeinen/Quasar/eventq"
	"strings"
	"sync"
)

var IdAssigned = eventq.NewEventType()

type idAssigner struct {
	names         []string
	mapping       map[string]Id
	constantNames []string
	lock          sync.RWMutex

	eventq.Messenger
}
type Id int

func newIdAssigner(constantNames ...string) idAssigner {
	ret := idAssigner{
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
	this.names = make([]string, 0, 64)
	this.mapping = make(map[string]Id, 64)
	this.lock.Unlock()
	this.assign(append([]string{"Unknown"}, this.constantNames...))
}

func (this *idAssigner) assign(names []string) (ids []Id, added []bool) {
	this.lock.Lock()
	defer this.lock.Unlock()

	for _, name := range names {
		lname := strings.ToLower(name)
		_, exists := this.mapping[lname]
		if !exists {
			this.names = append(this.names, name)
			id := Id(len(this.names) - 1)
			this.mapping[lname] = id
			this.Event(IdAssigned, id)
		}
		added = append(added, !exists)
		ids = append(ids, this.mapping[lname])
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
	return this.names[id]
}
