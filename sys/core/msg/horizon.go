package msg

import (
	cbor "github.com/fxamacker/cbor/v2"
)

type Horizon map[string]uint64

func EncodeHorizon(h *Horizon) ([]byte, error) {
	b, err := cbor.Marshal(h)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func DecodeHorizon(b []byte) (*Horizon, error) {
	h := new(Horizon)
	err := cbor.Unmarshal(b, &h)
	if err != nil {
		return h, err
	}
	return h, nil
}

func (h *Horizon) Update(device string, seq uint64) {
	(*h)[device] = seq
}
