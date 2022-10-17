package slate

import "slater/core/msg"

type Slate interface {
	Name() string
	Write(*msg.Message) error
	On(string, func(*msg.Message))
	Once(string, func(*msg.Message))
	Get(uint64) (*msg.Message, error)
	GetRange(int, int) ([]*msg.Message, error)
	Count() uint64
}
