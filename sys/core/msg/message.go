package msg

import (
	cbor "github.com/fxamacker/cbor/v2"
	"time"
)

type Message struct {
	Slate   string
	User    string
	Device  string
	Seq     uint64
	Sent    int64
	Prev    string
	Next    string
	Kind    string
	Event   string
	Content map[string]any
}

// Using fxmacker's defaults for now...
// (there was something I wanted to configure later on, but can't remember what right now...)

func Encode(m *Message) ([]byte, error) {
	b, err := cbor.Marshal(m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func Decode(b []byte) (*Message, error) {
	m := new(Message)
	err := cbor.Unmarshal(b, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func Timestamp() int64 {
	return time.Now().UnixMilli()
}
