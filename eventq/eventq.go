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
	eventTypeCounter uint32
	actionIdCounter  actionId
	queue            = make(map[EventType][]eventAction, 32)
	lock             sync.RWMutex
)

func NewEventType() EventType {
	return EventType(atomic.AddUint32(&eventTypeCounter, 1))
}

func On(event EventType) doWhat {
	return doWhat(event)
}
func (this doWhat) Do(what func(...interface{})) ActionHandle {
	lock.Lock()
	defer lock.Unlock()

	event := EventType(this)
	newId := actionId(atomic.AddUint64((*uint64)(&actionIdCounter), 1))
	queue[event] = append(queue[event], eventAction{newId, what})

	return ActionHandle{event, newId}
}

func Event(event EventType, args ...interface{}) {
	lock.RLock()
	actions := queue[event]
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
	if atomic.LoadUint32((*uint32)(&this.event)) == 0 {
		return //Already cancelled
	}

	lock.Lock()
	defer lock.Unlock()

	idx := -1
	actions := queue[this.event]
	for i, action := range actions {
		if action.id == this.id {
			idx = i
			break
		}
	}

	if idx != -1 {
		actions, actions[len(actions)-1] = append(actions[:idx], actions[idx+1:]...), eventAction{0, nil} //delete, set the last elem to nil to prevent a memory leak
		queue[this.event] = actions

		this.event = EventType(0)
	}
}
