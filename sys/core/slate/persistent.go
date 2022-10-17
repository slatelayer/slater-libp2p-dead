package slate

import (
	"context"
	"errors"
	"golang.org/x/exp/slices"
	"strconv"

	cbor "github.com/fxamacker/cbor/v2"
	ds "github.com/ipfs/go-datastore"
	logging "github.com/ipfs/go-log/v2"

	"slater/core/msg"
	"slater/core/store"
)

const (
	ROOT    = "s"
	SEQ     = "sq"
	SUBLOGS = "sl"
)

var (
	ctx = context.TODO()
	log = logging.Logger("slater:core")
)

type PersistentSlate struct {
	name    string
	Device  string
	Store   store.Store
	Emitter *Emitter
}

func NewPersistentSlate(name, device string, db store.Store) *PersistentSlate {
	return &PersistentSlate{
		name:    name,
		Device:  device,
		Store:   db,
		Emitter: NewEmitter(),
	}
}

func (slate *PersistentSlate) Name() string {
	return slate.name
}

// record a message to this device's log, to then be replicated
// (derives Seq and Prev from current state)
func (slate *PersistentSlate) Send(m *msg.Message) error {
	m.Slate = slate.name
	m.Device = slate.Device

	sublog := slate.Device

	txn, err := slate.Store.Store.NewTransaction(ctx, false)
	if err != nil {
		return err
	}

	// NEXT: actually, save the horizon instead...
	sublogsNs := []string{ROOT, slate.name, SUBLOGS}
	sublogsKey := ds.KeyWithNamespaces(sublogsNs)
	sublogsBytes, err := txn.Get(ctx, sublogsKey)
	if err != nil {
		return err
	}
	var sublogs []string
	err = cbor.Unmarshal(sublogsBytes, sublogs)
	if err != nil {
		return err
	}

	if !slices.Contains(sublogs, sublog) {
		// lazy add, assuming more ceremony at a higher level to govern what sublogs are valid
		sublogs = append(sublogs, sublog)
		sublogsBytes, err = cbor.Marshal(sublogs)
		if err != nil {
			return err
		}
		txn.Put(ctx, sublogsKey, sublogsBytes)
	}

	logNs := []string{ROOT, slate.name, slate.Device}

	seqNs := append(logNs, SEQ)
	seqKey := ds.KeyWithNamespaces(seqNs)

	seqBytes, err := txn.Get(ctx, seqKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	var seq uint64
	err = cbor.Unmarshal(seqBytes, seq)
	if err != nil {
		return err
	}

	seq++
	m.Seq = seq

	seqBytes, err = cbor.Marshal(seq)
	if err != nil {
		return err
	}

	txn.Put(ctx, seqKey, seqBytes)

	lastMessageNs := []string{ROOT, slate.name}
	lastMessageKey := ds.KeyWithNamespaces(lastMessageNs)
	lastMessageBytes, err := txn.Get(ctx, lastMessageKey)
	if err != nil && !errors.Is(err, ds.ErrNotFound) {
		return err
	}
	var lastMessage string
	err = cbor.Unmarshal(lastMessageBytes, lastMessage)
	if err != nil {
		return err
	}

	m.Prev = lastMessage

	msgNs := append(logNs, strconv.FormatUint(seq, 10))
	msgKey := ds.KeyWithNamespaces(msgNs)

	msgBytes, err := msg.Encode(m)
	if err != nil {
		return err
	}

	txn.Put(ctx, msgKey, msgBytes)

	lastMessageBytes, err = cbor.Marshal(msgKey.String())
	txn.Put(ctx, lastMessageKey, lastMessageBytes)

	horizon := new(msg.Horizon)
	for log := range sublogs {

	}

	txn.Commit(ctx)

	slate.Emitter.Emit(m)

	return nil
}

// record a message which was written on another device
// (expects Seq and Prev fields to be written already)
func (slate *PersistentSlate) Recv(m *msg.Message) error {
	txn, err := slate.Store.Store.NewTransaction(ctx, false)
	if err != nil {
		return err
	}

	msgNs := []string{ROOT, slate.Name(), m.Device, strconv.FormatUint(m.Seq, 10)}
	msgKey := ds.KeyWithNamespaces(msgNs)
	msgKeyStr := msgKey.String()

	prevKey := ds.NewKey(m.Prev)
	prevMsgBytes, err := txn.Get(ctx, prevKey)
	if err != nil {
		if errors.Is(err, ds.ErrNotFound) {
			// we don't have that message yet!
			// XXX this leaves a gap, and I'm not sure what to do about it yet!
			// ...should this even happen? hopefully we can avoid this completely.
			// we want to always append, never insert...
			log.Debug("missing link")
		} else {
			return err
		}
	}

	prevMsg, err := msg.Decode(prevMsgBytes)
	if err != nil {
		return err
	}

	other := prevMsg.Next
	if other != "" {
		otherKey := ds.NewKey(other)
		otherMsgBytes, err := txn.Get(ctx, otherKey)
		if err != nil {
			return err
		}
		otherMsg, err := msg.Decode(otherMsgBytes)
		if err != nil {
			return err
		}

		if comesBefore(m, otherMsg) {
			m.Next = other
			otherMsg.Prev = msgKeyStr

			prevMsg.Next = msgKeyStr
			prevRec, err := msg.Encode(prevMsg)
			if err != nil {
				return err
			}
			txn.Put(ctx, prevKey, prevRec)
		} else {
			m.Prev = other
			otherMsg.Next = msgKeyStr
		}

		otherRec, err := msg.Encode(otherMsg)
		if err != nil {
			return err
		}
		txn.Put(ctx, otherKey, otherRec)
	} else {
		prevMsg.Next = msgKeyStr
		prevRec, err := msg.Encode(prevMsg)
		if err != nil {
			return err
		}
		txn.Put(ctx, prevKey, prevRec)
	}

	rec, err := msg.Encode(m)
	if err != nil {
		return err
	}

	txn.Put(ctx, msgKey, rec)

	txn.Commit(ctx)

	slate.Emitter.Emit(m)

	return nil
}

func comesBefore(a, b *msg.Message) bool {
	if a.Sent == b.Sent {
		return a.Device < b.Device
	}
	return a.Sent < b.Sent
}

func (slate *PersistentSlate) On(kind string, fn func(*msg.Message)) {
	slate.Emitter.On(kind, fn)
}

func (slate *PersistentSlate) Once(kind string, fn func(*msg.Message)) {
	slate.Emitter.Once(kind, fn)
}
