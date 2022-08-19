package core

import (
	cbor "github.com/fxamacker/cbor/v2"
	"time"
)

// NOTE: `message` will be strongly-typed and versioned once design stabilizes more...

type message map[string]any

// Using fxmacker's defaults for now...
// (there was something I wanted to configure later on, but can't remember what right now...)

func encode(msg message) ([]byte, error) {
	b, err := cbor.Marshal(msg)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func decode(b []byte) (message, error) {
	m := make(message)
	err := cbor.Unmarshal(b, &m)
	if err != nil {
		return m, err
	}
	return m, nil
}

func timestamp() int64 {
	return time.Now().UnixMilli()
}
