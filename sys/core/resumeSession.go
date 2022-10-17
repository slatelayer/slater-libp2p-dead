package core

import (
	"errors"

	"github.com/libp2p/go-libp2p-core/crypto"

	"slater/core/slate"
	"slater/core/store"
)

func resumeSession(core *Core, feed slate.Slate, say func(...string), sessions []string) (store.Store, *node) {
	createNew := "Create a New Session"
	choices := append(sessions, createNew)
	sessionName := <-chooseSession(feed, choices)

	if sessionName == createNew {
		return setupUser(core, feed, say)
	}

	var passphrase, pin string
	var db store.Store
	var err error
	for db.Store == nil {
		passphrase = <-promptPassphrase(feed)
		pin = <-promptPIN(feed)

		key, err := getMasterKey(core.root, sessionName, passphrase, pin)
		if err != nil {
			if errors.Is(err, errAuthFail) {
				say("😬 I could not confirm those credentials.")
				say("Take a deep breath...")
				say("Inhale...")
				say("Exhale...")
				say("And let's try again...")
				return resumeSession(core, feed, say, sessions)
			}

			if errors.Is(err, errLostSalt) {
				say("Hey, my salt file is missing! Now I can't decrypt your data.")
			} else if errors.Is(err, errLostHash) {
				say("Hey, my hash file is missing! Now I can't decrypt your data.")
			}

			say("Please don't delete my files. I hope you have a full replica of your data on another device!")
			say("If so, please make sure it's online and running session " + sessionName)

			store.RemoveStore(core.root, sessionName)

			key = createMasterKey(core.root, sessionName, passphrase, pin)
		}

		db, err = store.OpenStore(core.root, sessionName, key)

		if err != nil {
			if len(sessions) > 1 {
				say("### Decryption failed. Please confirm your id and credentials.")

				chooseAnother := <-chooseAnotherSession(feed)

				if chooseAnother {
					return resumeSession(core, feed, say, sessions)
				} else {
					continue
				}
			} else {
				say("### Decryption failed. Please confirm your credentials.")

				tryAgain := <-promptTryAgain(feed)
				if !tryAgain {
					return resumeSession(core, feed, say, sessions)
				}
			}
		}
	}

	keyBytes, err := db.Get(KEYKEY)

	if err != nil {
		log.Panic(err)
	}

	privKey, err := crypto.UnmarshalPrivateKey(keyBytes)

	if err != nil {
		log.Panic(err)
	}

	node, err := startNet(privKey, db)

	log.Debug("node: ", node.host.ID())

	if err != nil {
		log.Panic(err)
	}

	connect(core, db, node, sessionName, passphrase, pin)

	return db, node
}

func chooseSession(feed slate.Slate, sessions []string) chan string {
	return choose(feed, "setup:sessionID", "Start a session", sessions)
}

func promptPassphrase(feed slate.Slate) chan string {
	return promptSecret(feed, "setup:passphrase", "Enter your passphrase")
}

func promptPIN(feed slate.Slate) chan string {
	return promptSecret(feed, "setup:pin", "Enter your PIN")
}

func chooseAnotherSession(feed slate.Slate) chan bool {
	return affirm(feed, "setup:chooseAnotherSession?",
		"## Do you want to choose a different session?", "Yes", "No, let's try again")
}

func promptTryAgain(feed slate.Slate) chan bool {
	return affirm(feed, "setup:tryAgain?",
		"## Do you want to try again?", "Yes", "No")
}
