package core

import (
	"context"
	"crypto/ed25519"
	"errors"
	"golang.org/x/exp/slices"
	"strings"
	"time"

	"github.com/fxamacker/cbor/v2"
	"github.com/libp2p/go-libp2p-core/crypto"
	_peer "github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"lukechampine.com/frand"

	"slater/core/msg"
	"slater/core/slate"
	"slater/core/store"
)

const (
	DELAY_MIN = 0.25
	DELAY_MAX = 1.25

	KEYKEY     = "k"
	DEVICESKEY = "d"
)

func wait() {
	min := DELAY_MIN
	max := DELAY_MAX
	d := (max-min)*frand.Float64() + min
	<-time.After(time.Duration(d) * time.Second)
}

func runSetup(core *Core, feed slate.Slate) (store.Store, *node) {
	text := func(body string) {
		feed.Write(&msg.Message{
			Slate: "setup",
			User:  "system",
			Kind:  "text",
			Sent:  msg.Timestamp(),
			Content: map[string]any{
				"body": body,
			},
		})
	}

	say := func(things ...string) { wait(); text(strings.Join(things, "\n")) }

	// TODO instead of saying "hello" in English,
	// we will say it in many languages, as a language chooser.
	// so, check if any languages have been saved first, and
	// if not, we'll show this big wall of greetings,
	// and when you click on one, it will send a reply in that language and save the setting.
	// the user may also type a reply. in that case, it should fade out all the greetings
	// except for the ones we think the user may be typing, and then
	// if there is only one, sending the message should select it.
	// else if there are more than one, it should be obvious that the user should tap one to select it.

	text("# Hello!")

	stores, err := store.FindStores(core.root)
	if err != nil {
		log.Fatal("serious problem with disk access")
	}

	if len(stores) == 0 {
		isNewUser := <-askIfNew(feed)

		if isNewUser {
			return setupUser(core, feed, say)

		} else {
			return setupDevice(core, feed, say)
		}
	}
	return resumeSession(core, feed, say, stores)
}

func askIfNew(feed slate.Slate) chan bool {
	return affirm(feed, "setup:newUser?", "## Are you a new user?",
		"Yes: setup a new ID", "No: setup this device")
}

func prompt(feed slate.Slate, evt string, body string) chan string {
	out := make(chan string, 0)

	feed.Write(&msg.Message{
		User: "system",
		Kind: "text",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body": body,
			"prompt": map[string]any{
				"event": evt,
				"kind":  "text",
			},
		},
	})

	feed.Once(evt, func(m *msg.Message) {
		str, there := m.Content["body"]
		if !there {
			log.Panic("missing body")
		}
		s, ok := str.(string)
		if !ok {
			log.Panic("wrong value type")
		}
		out <- s
	})

	return out
}

func secret(feed slate.Slate, label string, things ...string) {
	feed.Write(&msg.Message{
		User: "system",
		Kind: "secretText",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body":       label,
			"secretText": strings.Join(things, "\n\n"),
		},
	})
}

func promptSecret(feed slate.Slate, evt string, body string) chan string {
	out := make(chan string, 0)

	feed.Write(&msg.Message{
		User: "system",
		Kind: "text",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body": body,
			"prompt": map[string]any{
				"event": evt,
				"kind":  "secretText",
			},
		},
	})

	feed.Once(evt, func(m *msg.Message) {
		str, there := m.Content["secretText"]
		if !there {
			log.Panic("missing secretText")
		}
		s, ok := str.(string)
		if !ok {
			log.Panic("wrong value type")
		}
		out <- s
	})

	return out
}

func choose(feed slate.Slate, evt string, body string, choices []string) chan string {
	out := make(chan string, 0)

	feed.Write(&msg.Message{
		User: "system",
		Kind: "text",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body": body,
			"prompt": map[string]any{
				"event":   evt,
				"kind":    "choice",
				"choices": choices,
			},
		},
	})

	feed.Once(evt, func(m *msg.Message) {
		fl, there := m.Content["choice"]
		if !there {
			log.Panic("missing choice")
		}
		f, ok := fl.(float64)
		if !ok {
			log.Panic("wrong value type")
		}
		i := int64(f)
		val := choices[i]
		out <- val
	})

	return out
}

func affirm(feed slate.Slate, evt string, body string, choices ...string) chan bool {
	out := make(chan bool, 0)

	feed.Write(&msg.Message{
		User: "system",
		Kind: "text",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body": body,
			"prompt": map[string]any{
				"event":   evt,
				"kind":    "choice",
				"choices": choices,
			},
		},
	})

	feed.Once(evt, func(m *msg.Message) {
		fl, there := m.Content["choice"]
		if !there {
			log.Panic("missing choice")
		}
		f, ok := fl.(float64)
		if !ok {
			log.Panic("wrong value type")
		}

		choice := int64(f)
		if choice == 0 {
			out <- true
		} else {
			out <- false
		}
	})

	return out
}

func affirmSecret(feed slate.Slate, evt string, body string, secret string, choices ...string) chan bool {
	out := make(chan bool, 0)

	feed.Write(&msg.Message{
		User: "system",
		Kind: "secretText",
		Sent: msg.Timestamp(),
		Content: map[string]any{
			"body":       body,
			"secretText": secret,
			"prompt": map[string]any{
				"event":   evt,
				"kind":    "choice",
				"choices": choices,
			},
		},
	})

	feed.Once(evt, func(m *msg.Message) {
		fl, there := m.Content["choice"]
		if !there {
			log.Panic("missing choice")
		}
		f, ok := fl.(float64)
		if !ok {
			log.Error("wrong value type")
		}
		choice := int64(f)
		if choice == 0 {
			out <- true
		} else {
			out <- false
		}
	})

	return out
}

func completeSetup(core *Core, name, passphrase, pin string) (store.Store, *node) {
	key := createMasterKey(core.root, name, passphrase, pin)

	db, err := store.OpenStore(core.root, name, key)

	if err != nil {
		log.Panic(err)
	}

	privKey, _, err := crypto.GenerateEd25519Key(frand.Reader)

	if err != nil {
		log.Panic(err)
	}

	keyBytes, err := crypto.MarshalPrivateKey(privKey)
	if err != nil {
		log.Panic(err)
	}

	err = db.Put([]string{KEYKEY}, keyBytes)
	if err != nil {
		log.Panic(err)
	}

	peer, err := startNet(privKey, db)

	if err != nil {
		log.Panic(err)
	}

	connect(core, db, peer, name, passphrase, pin)

	log.Debugf("discoverykey: %v, signet: %v", peer.discoveryKey, peer.signet)

	return db, peer
}

func connect(core *Core, db store.Store, peer *node, sessionName, passphrase, pin string) {
	devicesBytes, err := db.Get(DEVICESKEY)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			log.Panic(err)
		}
	} else {
		err = cbor.Unmarshal(devicesBytes, &core.devices)
		if err != nil {
			log.Panic(err)
		}
	}

	discoKey := discoveryKey(sessionName, passphrase, pin)
	peer.discoveryKey = discoKey

	signKey, err := deriveSignatureKey(sessionName, passphrase, pin)
	if err != nil {
		log.Panic(err)
	}

	signet := ed25519.Sign(signKey, []byte(peer.host.ID().String()))
	peer.signet = signet

	validator := func(ctx context.Context, pid _peer.ID, pmsg *pubsub.Message) pubsub.ValidationResult {
		p := pid.String()

		if p == peer.host.ID().String() {
			return pubsub.ValidationAccept
		}

		if slices.Contains(core.devices, p) {
			return pubsub.ValidationAccept
		}

		m, err := msg.Decode(pmsg.Data)
		if err != nil {
			log.Panic(err)
		}

		signet, there := m.Content["signet"]
		if !there {
			return pubsub.ValidationReject
		}

		signetBytes, ok := signet.([]byte)

		if !ok {
			log.Panic("invalid signet")
		}

		pubKey := signKey.Public().(ed25519.PublicKey)
		valid := ed25519.Verify(pubKey, []byte(p), signetBytes)

		if valid {
			core.devices = append(core.devices, p)
			bytes, err := cbor.Marshal(core.devices)
			if err != nil {
				log.Panic(err)
			}
			db.Put([]string{DEVICESKEY}, bytes)
			return pubsub.ValidationAccept
		}

		return pubsub.ValidationReject
	}

	peer.join(peer.discoveryKey, validator)

	go func() {
		for {
			peers := peer.psub.ListPeers(peer.discoveryKey)
			count := len(peers)
			if count > 0 {
				peer.send(peer.discoveryKey, &msg.Message{
					Slate: "setup",
					Kind:  "signet",
					Content: map[string]any{
						"signet": peer.signet,
					},
				})
				return
				// ...or should we continue sending periodically? 🤔...
				// Should we continue polling, and watch for more peers?
				// Or is there a more efficient means of detecting new peers?
				// TODO
			}
			time.Sleep(1 * time.Second)
		}
	}()
}
