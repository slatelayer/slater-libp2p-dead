package core

import (
	"errors"
	"sync"
)

type slate interface {
	id() string
	write(message) (int, error)
	listen(string, listener)
	on(string, handler)
	once(string, handler)
	get(int) (message, error)
	getRange(int, int) ([]message, error)
	count() int
}

type listener func(slate, message, int) bool
type handler func(slate, message, int)

const ALL = "*" //wildcard kind

type ephemeralSlate struct {
	name      string
	log       []message
	listeners map[string][]listener
	logLock   *sync.RWMutex
	emLock    *sync.RWMutex
}

func newEphemeralSlate(name string) *ephemeralSlate {
	return &ephemeralSlate{
		name:      name,
		log:       make([]message, 0),
		listeners: map[string][]listener{ALL: make([]listener, 0)},
		logLock:   &sync.RWMutex{},
		emLock:    &sync.RWMutex{},
	}
}

func (ephem *ephemeralSlate) id() string {
	return ephem.name
}

func (ephem *ephemeralSlate) write(msg message) (int, error) {
	//log.Debugf("writing message: %v", msg)
	count := ephem.count()

	kind, hasKind := msg["kind"]
	if !hasKind {
		return count, errors.New("invalid message: no 'kind'") // TODO error types
	}

	kindString, ok := kind.(string)
	if !ok {
		return count, errors.New("invalid message: kind not a string.")
	}

	event, hasEvent := msg["event"]
	var eventString string

	if hasEvent {
		eventString, ok = event.(string)
		if !ok {
			return count, errors.New("invalid event: kind not a string.")
		}
	}

	// First we append to the log,
	ephem.logLock.Lock()
	ephem.log = append(ephem.log, msg)
	ephem.logLock.Unlock()

	// Then we notify its listeners:
	ephem.emLock.RLock()

	wildcardListeners, _ := ephem.listeners[ALL]

	messageListeners, haveMessageListeners := ephem.listeners[kindString]

	var eventListeners []listener
	var haveEventListeners bool
	if hasEvent {
		eventListeners, haveEventListeners = ephem.listeners[eventString]
	}

	ephem.emLock.RUnlock()

	wl2rm := make([]int, 0)
	for i, fn := range wildcardListeners {
		if fn == nil {
			continue
		}
		if !fn(ephem, msg, count) {
			wl2rm = append(wl2rm, i)
		}
	}

	ml2rm := make([]int, 0)
	if haveMessageListeners {
		for i, fn := range messageListeners {
			if fn == nil {
				continue
			}
			if !fn(ephem, msg, count) {
				ml2rm = append(ml2rm, i)
			}
		}
	}

	el2rm := make([]int, 0)
	if hasEvent && haveEventListeners {
		for i, fn := range eventListeners {
			if fn == nil {
				continue
			}
			if !fn(ephem, msg, count) {
				el2rm = append(el2rm, i)
			}
		}
	}

	if len(wl2rm) > 0 || len(ml2rm) > 0 || len(el2rm) > 0 {
		ephem.emLock.Lock()

		for _, idx := range wl2rm {
			ephem.listeners[ALL][idx] = nil
		}

		for _, idx := range ml2rm {
			ephem.listeners[kindString][idx] = nil
		}

		for _, idx := range el2rm {
			ephem.listeners[eventString][idx] = nil
		}

		ephem.emLock.Unlock()
	}

	return count + 1, nil
}

func (ephem *ephemeralSlate) listen(kind string, fn listener) {
	ephem.emLock.Lock()
	defer ephem.emLock.Unlock()

	listeners, there := ephem.listeners[kind]

	if there {
		ephem.listeners[kind] = append(listeners, fn)
	} else {
		ephem.listeners[kind] = []listener{fn}
	}
}

func (ephem *ephemeralSlate) on(kind string, h handler) {
	fn := func(s slate, msg message, idx int) bool {
		h(s, msg, idx)
		return true
	}

	ephem.listen(kind, fn)
}

func (ephem *ephemeralSlate) once(kind string, h handler) {
	fn := func(s slate, msg message, idx int) bool {
		h(s, msg, idx)
		return false
	}

	ephem.listen(kind, fn)
}

func (ephem *ephemeralSlate) get(idx int) (message, error) {
	ephem.logLock.RLock()
	defer ephem.logLock.RUnlock()

	count := len(ephem.log)

	if idx >= count {
		return message{}, errors.New("ephem.get: index out of bounds!")
	}

	return ephem.log[idx], nil
}

func (ephem *ephemeralSlate) getRange(from int, including int) ([]message, error) {
	ephem.logLock.RLock()
	defer ephem.logLock.RUnlock()

	if including == -1 {
		return ephem.log[from:], nil
	}

	high := including + 1
	if high >= len(ephem.log) {
		return []message{}, errors.New("ephem.range: range exceeded bounds!")
	}

	return ephem.log[from:high], nil
}

func (ephem *ephemeralSlate) count() int {
	ephem.logLock.RLock()
	defer ephem.logLock.RUnlock()

	return len(ephem.log)
}
