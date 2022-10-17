package slate

import (
	"golang.org/x/exp/slices"
	"sync"

	"slater/core/msg"
)

type Emitter struct {
	listeners map[string][]listener
	lock      *sync.Mutex
}

type listener struct {
	fn      func(*msg.Message)
	persist bool
}

const ALL = "*" //wildcard kind

func NewEmitter() *Emitter {
	return &Emitter{
		listeners: map[string][]listener{ALL: make([]listener, 0)},
		lock:      &sync.Mutex{},
	}
}

func (emitter *Emitter) On(kind string, fn func(*msg.Message)) {
	emitter.listen(kind, fn, true)
}

func (emitter *Emitter) Once(kind string, fn func(*msg.Message)) {
	emitter.listen(kind, fn, false)
}

func (emitter *Emitter) listen(kind string, fn func(*msg.Message), persist bool) {
	emitter.lock.Lock()
	listeners, there := emitter.listeners[kind]
	if there {
		listeners = append(listeners, listener{fn, persist})
	}
}

func (emitter *Emitter) Emit(m *msg.Message) {
	kind := m.Kind
	event := m.Event

	emitter.lock.Lock()

	wildcardListeners := emitter.listeners[ALL]

	kindListeners, haveKindListeners := emitter.listeners[kind]
	eventListeners, haveEventListeners := emitter.listeners[event]

	for i, h := range wildcardListeners {
		go h.fn(m)
		if !h.persist {
			emitter.listeners[ALL] = slices.Delete(emitter.listeners[ALL], i, i)
		}
	}

	if haveKindListeners {
		for i, h := range kindListeners {
			go h.fn(m)
			if !h.persist {
				emitter.listeners[kind] = slices.Delete(emitter.listeners[kind], i, i)
			}
		}
	}

	if haveEventListeners {
		for i, h := range eventListeners {
			go h.fn(m)
			if !h.persist {
				emitter.listeners[event] = slices.Delete(emitter.listeners[event], i, i)
			}
		}
	}

	emitter.lock.Unlock()
}
