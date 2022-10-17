package slate

import (
	"errors"
	"sync"

	"slater/core/msg"
)

type EphemeralSlate struct {
	name    string
	Log     []*msg.Message
	Lock    *sync.RWMutex
	Emitter *Emitter
}

func NewEphemeralSlate(name string) *EphemeralSlate {
	return &EphemeralSlate{
		name:    name,
		Log:     make([]*msg.Message, 0),
		Lock:    &sync.RWMutex{},
		Emitter: NewEmitter(),
	}
}

func (slate *EphemeralSlate) Name() string {
	return slate.name
}

func (slate *EphemeralSlate) Write(m *msg.Message) error {
	m.Slate = slate.Name()
	slate.Lock.Lock()
	slate.Log = append(slate.Log, m)
	slate.Lock.Unlock()

	slate.Emitter.Emit(m)

	return nil
}

func (slate *EphemeralSlate) On(kind string, fn func(*msg.Message)) {
	slate.Emitter.On(kind, fn)
}

func (slate *EphemeralSlate) Once(kind string, fn func(*msg.Message)) {
	slate.Emitter.Once(kind, fn)
}

func (slate *EphemeralSlate) Get(idx uint64) (*msg.Message, error) {
	slate.Lock.RLock()
	defer slate.Lock.RUnlock()

	count := uint64(len(slate.Log))

	if idx >= count {
		return nil, errors.New("slate.get: index out of bounds!")
	}

	return slate.Log[idx], nil
}

func (slate *EphemeralSlate) GetRange(from, including int) ([]*msg.Message, error) {
	slate.Lock.RLock()
	defer slate.Lock.RUnlock()

	high := including + 1
	if uint64(high) >= uint64(len(slate.Log)) {
		return nil, errors.New("slate.range: range exceeded bounds!")
	}

	return slate.Log[from:high], nil
}

func (slate *EphemeralSlate) Count() uint64 {
	slate.Lock.RLock()
	defer slate.Lock.RUnlock()

	return uint64(len(slate.Log))
}
