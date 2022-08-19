package core

import (
	"errors"
	"github.com/libp2p/go-libp2p-core/crypto"
)

func resumeSession(core *Core, feed slate, say func(...string), sessions []string) (datastore, node) {
	createNew := "Create a New Session"
	choices := append(sessions, createNew)
	sessionName := <-chooseSession(feed, choices)

	if sessionName == createNew {
		return setupUser(core, feed, say)
	}

	var passphrase, pin string
	var store datastore
	var err error
	for store.store == nil {
		passphrase = <-promptPassphrase(feed)
		pin = <-promptPIN(feed)

		key, err := getMasterKey(sessionName, passphrase, pin)
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

			deleteStore(sessionName)

			key = createMasterKey(sessionName, passphrase, pin)
		}

		store, err = openStore(sessionName, key)

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

	keyBytes, err := store.get(KEYKEY)

	if err != nil {
		log.Panic(err)
	}

	privKey, err := crypto.UnmarshalPrivateKey(keyBytes)

	if err != nil {
		log.Panic(err)
	}

	node, err := startNet(privKey, store)

	log.Debug("node: ", node.host.ID())

	if err != nil {
		log.Panic(err)
	}

	connect(core, store, node, sessionName, passphrase, pin)

	return store, node
}

func chooseSession(feed slate, sessions []string) chan string {
	return choose(feed, "setup:sessionID", "Start a session", sessions)
}

func promptPassphrase(feed slate) chan string {
	return promptSecret(feed, "setup:passphrase", "Enter your passphrase")
}

func promptPIN(feed slate) chan string {
	return promptSecret(feed, "setup:pin", "Enter your pin")
}

func chooseAnotherSession(feed slate) chan bool {
	return affirm(feed, "setup:chooseAnotherSession?",
		"## Do you want to choose a different session?", "Yes", "No, let's try again")
}

func promptTryAgain(feed slate) chan bool {
	return affirm(feed, "setup:tryAgain?",
		"## Do you want to try again?", "Yes", "No")
}
