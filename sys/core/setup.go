package core

import (
	"context"
	"crypto/ed25519"
	"errors"
	"github.com/fxamacker/cbor/v2"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-pubsub"
	"golang.org/x/exp/slices"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p-core/crypto"
	"lukechampine.com/frand"
)

const (
	DELAY_MIN = 0.25
	DELAY_MAX = 1.25
)

func wait() {
	min := DELAY_MIN
	max := DELAY_MAX
	d := (max-min)*frand.Float64() + min
	<-time.After(time.Duration(d) * time.Second)
}

func runSetup(core *Core, feed slate) (datastore, node) {
	text := func(body string) {
		feed.write(message{
			"slate":  "setup",
			"author": "system",
			"kind":   "text",
			"body":   body,
			"sent":   timestamp(),
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

	stores, err := findStores()
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

func askIfNew(feed slate) chan bool {
	return affirm(feed, "setup:newUser?", "## Are you a new user?",
		"Yes: setup a new ID", "No: setup this device")
}

func prompt(feed slate, evt string, body string) chan string {
	out := make(chan string, 0)

	feed.write(message{
		"slate":  feed.id(),
		"author": "system",
		"kind":   "text",
		"body":   body,
		"prompt": message{
			"event": evt,
			"kind":  "text",
		},
		"sent": timestamp(),
	})

	feed.once(evt, func(feed slate, msg message, idx int) {
		s, ok := msg["body"].(string)
		if ok {
			out <- s
		} else {
			log.Error("wrong value type")
		}
	})

	return out
}

func secret(feed slate, label string, things ...string) {
	feed.write(message{
		"slate":      feed.id(),
		"author":     "system",
		"kind":       "secretText",
		"body":       label,
		"secretText": strings.Join(things, "\n\n"),
		"sent":       timestamp(),
	})
}

func promptSecret(feed slate, evt string, body string) chan string {
	out := make(chan string, 0)

	feed.write(message{
		"slate":  feed.id(),
		"author": "system",
		"kind":   "text",
		"body":   body,
		"prompt": message{
			"event": evt,
			"kind":  "secretText",
		},
		"sent": timestamp(),
	})

	feed.once(evt, func(feed slate, msg message, idx int) {
		s, ok := msg["secretText"].(string)
		if ok {
			out <- s
		} else {
			log.Error("wrong value type")
		}
	})

	return out
}

func choose(feed slate, evt string, body string, choices []string) chan string {
	out := make(chan string, 0)

	feed.write(message{
		"slate":  feed.id(),
		"author": "system",
		"kind":   "text",
		"body":   body,
		"prompt": message{
			"event":   evt,
			"kind":    "choice",
			"choices": choices,
		},
		"sent": timestamp(),
	})

	feed.once(evt, func(feed slate, msg message, idx int) {
		f, ok := msg["choice"].(float64)
		if ok {
			i := int64(f)
			val := choices[i]
			out <- val
		} else {
			log.Error("wrong value type")
		}
	})

	return out
}

func affirm(feed slate, evt string, body string, choices ...string) chan bool {
	out := make(chan bool, 0)

	feed.write(message{
		"slate":  feed.id(),
		"author": "system",
		"kind":   "text",
		"body":   body,
		"prompt": message{
			"event":   evt,
			"kind":    "choice",
			"choices": choices,
		},
		"sent": timestamp(),
	})

	feed.once(evt, func(feed slate, msg message, idx int) {
		f, ok := msg["choice"].(float64)
		if ok {
			choice := int64(f)
			if choice == 0 {
				out <- true
			} else {
				out <- false
			}
		} else {
			log.Error("wrong value type")
		}
	})

	return out
}

func affirmSecret(feed slate, evt string, body string, secret string, choices ...string) chan bool {
	out := make(chan bool, 0)

	feed.write(message{
		"slate":      feed.id(),
		"author":     "system",
		"kind":       "secretText",
		"body":       body,
		"secretText": secret,
		"prompt": message{
			"event":   evt,
			"kind":    "choice",
			"choices": choices,
		},
		"sent": timestamp(),
	})

	feed.once(evt, func(feed slate, msg message, idx int) {
		f, ok := msg["choice"].(float64)
		if ok {
			choice := int64(f)
			if choice == 0 {
				out <- true
			} else {
				out <- false
			}
		} else {
			log.Error("wrong value type")
		}
	})

	return out
}

func completeSetup(core *Core, name, passphrase, pin string) (datastore, node) {
	key := createMasterKey(name, passphrase, pin)

	store, err := openStore(name, key)

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

	err = store.put(KEYKEY, keyBytes)
	if err != nil {
		log.Panic(err)
	}

	peer, err := startNet(privKey, store)

	if err != nil {
		log.Panic(err)
	}

	connect(core, store, peer, name, passphrase, pin)

	return store, peer
}

func connect(core *Core, store datastore, host node, sessionName, passphrase, pin string) {
	devicesBytes, err := store.get(DEVICESKEY)
	if err != nil {
		if !errors.Is(err, ErrNotFound) {
			log.Panic(err)
		}
	}

	err = cbor.Unmarshal(devicesBytes, &core.devices)
	if err != nil {
		log.Panic(err)
	}

	discoKey := discoveryKey(sessionName, passphrase, pin)
	signKey, err := deriveSignatureKey(sessionName, passphrase, pin)
	if err != nil {
		log.Panic(err)
	}

	host.discoveryKey = discoKey
	host.signet = ed25519.Sign(signKey, []byte(host.host.ID().String()))

	validator := func(ctx context.Context, pid peer.ID, msg *pubsub.Message) pubsub.ValidationResult {
		p := pid.String()

		if p == host.host.ID().String() {
			return pubsub.ValidationAccept
		}

		if slices.Contains(core.devices, p) {
			return pubsub.ValidationAccept
		}

		m, err := decode(msg.Data)
		if err != nil {
			log.Panic(err)
		}

		signet, there := m["signet"]
		if !there {
			return pubsub.ValidationReject
		}

		pubKey := signKey.Public().(ed25519.PublicKey)
		valid := ed25519.Verify(pubKey, []byte(p), signet.([]byte))

		if valid {
			core.devices = append(core.devices, p)
			bytes, err := cbor.Marshal(core.devices)
			if err != nil {
				log.Panic(err)
			}
			store.put(DEVICESKEY, bytes)
			return pubsub.ValidationAccept
		}

		return pubsub.ValidationReject
	}

	host.join(discoKey, validator)

	go func() {
		for {
			peers := host.psub.ListPeers(discoKey)
			count := len(peers)
			if count > 0 {
				host.send(discoKey, message{
					"signet": host.signet,
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
