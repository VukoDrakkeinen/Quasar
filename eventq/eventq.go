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
	messenger *Messenger
	event     EventType
	id        actionId
}
type doWhat struct {
	messenger *Messenger
	event     EventType
}

var (
	eventTypeCounter uint32
	actionIdCounter  actionId
	global           = NewMessenger()
)

func NewEventType() EventType {
	return EventType(atomic.AddUint32(&eventTypeCounter, 1))
}

type Messenger struct {
	receivers map[EventType][]eventAction
	lock      sync.RWMutex
}

func NewMessenger() Messenger {
	return Messenger{
		receivers: make(map[EventType][]eventAction, 32),
	}
}

func (this *Messenger) On(event EventType) doWhat {
	return doWhat{this, event}
}

func (this doWhat) Do(what func(...interface{})) ActionHandle {
	messenger := this.messenger
	messenger.lock.Lock()
	defer messenger.lock.Unlock()

	event := this.event
	newId := actionId(atomic.AddUint64((*uint64)(&actionIdCounter), 1))
	messenger.receivers[event] = append(messenger.receivers[event], eventAction{newId, what})

	return ActionHandle{messenger, event, newId}
}

func (this *Messenger) Event(event EventType, args ...interface{}) {
	this.lock.RLock()
	actions := this.receivers[event]
	this.lock.RUnlock()

	for _, action := range actions {
		func() {
			defer func() {
				if err := recover(); err != nil {
					qlog.Log(qlog.Error, qerr.Chain("Processing event failed", err.(error)))
					qlog.Logf(qlog.Error, "\n%s\n", qutils.Stack())
				}
			}()

			action.do(args...)
		}()
	}
}

func (this *ActionHandle) Cancel() {
	if atomic.SwapUint32((*uint32)(&this.event), 0) == 0 {
		return //Already cancelled
	}

	messenger := this.messenger
	messenger.lock.Lock()
	defer messenger.lock.Unlock()

	idx := -1
	actions := messenger.receivers[this.event]
	for i, action := range actions {
		if action.id == this.id {
			idx = i
			break
		}
	}

	if idx != -1 {
		actions, actions[len(actions)-1] = append(actions[:idx], actions[idx+1:]...), eventAction{0, nil} //delete then set the last elem to nil to prevent a memory leak
		messenger.receivers[this.event] = actions
	}
}

func On(event EventType) doWhat {
	return global.On(event)
}

func Event(event EventType, args ...interface{}) {
	global.Event(event, args)
}
