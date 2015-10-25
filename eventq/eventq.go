package eventq

import (
	"github.com/VukoDrakkeinen/Quasar/datadir/qlog"
	"github.com/VukoDrakkeinen/Quasar/qutils"
	"github.com/VukoDrakkeinen/Quasar/qutils/qerr"
	"sync"
	"sync/atomic"
)

type EventType uint32
type actionId uint64
type eventAction struct {
	id actionId
	do func(...interface{})
}
type ActionHandle struct {
	event EventType
	id    actionId
}
type doWhat EventType

var (
	eventTypeCounter EventType
	actionIdCounter  actionId
	eq               = make(map[EventType][]eventAction, 32)
	lock             sync.RWMutex
)

func NewEventType() EventType {
	return EventType(atomic.AddUint32((*uint32)(&eventTypeCounter), 1))
}

func On(event EventType) doWhat {
	return doWhat(event)
}
func (this doWhat) Do(what func(...interface{})) ActionHandle {
	lock.Lock()
	defer lock.Unlock()

	event := EventType(this)
	newId := actionId(atomic.AddUint64((*uint64)(&actionIdCounter), 1))
	eq[event] = append(eq[event], eventAction{newId, what})

	return ActionHandle{event, newId}
}

func Event(event EventType, args ...interface{}) {
	lock.RLock()
	actions := eq[event]
	lock.RUnlock()

	defer func() {
		if err := recover(); err != nil {
			qlog.Log(qlog.Error, qerr.Chain("Processing event failed", err.(error)))
			qlog.Logf(qlog.Error, "\n%s\n", qutils.Stack())
		}
	}()
	for _, action := range actions {
		action.do(args...)
	}
}

func (this *ActionHandle) Cancel() {
	lock.Lock()
	defer lock.Unlock()

	idx := -1
	actions := eq[this.event]
	for i, action := range actions {
		if action.id == this.id {
			idx = i
			break
		}
	}

	if idx != -1 {
		actions, actions[len(actions)-1] = append(actions[:idx], actions[idx+1:]...), eventAction{0, nil} //delete, set the last elem to nil to prevent a memory leak
		eq[this.event] = actions
	}
}
