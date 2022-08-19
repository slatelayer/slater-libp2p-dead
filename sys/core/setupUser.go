package core

import (
	"time"
)

func setupUser(core *Core, feed slate, say func(...string)) (datastore, node) {
	say("Alright, I will create new random credentials, and I need you to **write them down** and *put them in your wallet*.\n",
		"So get ready to write, and make sure nobody else is looking at your screen!",
	)

	ready := <-askIfReady(feed)
	for !ready {
		wait()
		ready = <-askIfReadyNow(feed)
	}

	newName := generateSessionName()
	nameAccepted := <-proposeName(feed, newName)
	for !nameAccepted {
		newName = generateSessionName()
		nameAccepted = <-proposeAnotherName(feed, newName)
	}

	newPhrase := generatePassphrase()
	phraseAccepted := <-proposePhrase(feed, newPhrase)
	for !phraseAccepted {
		newPhrase = generatePassphrase()
		phraseAccepted = <-proposeAnotherPhrase(feed, newPhrase)
	}

	say("You'll also need a PIN number.")

	newPin := generatePin()
	pinAccepted := <-proposePin(feed, newPin)
	for !pinAccepted {
		newPin = generatePin()
		pinAccepted = <-proposePin(feed, newPin)
	}

	say("Awesome. Now write it down.",
		"Write on one sheet of paper so it doesn't imprint on another.",
		"Make sure nobody is looking!",
	)

	say("Write today's date on it, so when we make a new one it will be easy to see that this one is older.")

	now := time.Now()

	secret(feed, now.Format("2006-01-02"), newName, newPhrase, newPin)

	writtenDown := <-promptWrittenDown(feed)
	for !writtenDown {
		wait()
		writtenDown = <-promptWrittenDownAgain(feed)
	}

	say("Put it with your money.",
		"Make a copy, and put it in a safe or something.\n",
		"Make another copy, and put it in your lawyer's safe or something.",
	)

	say("_Don't lose it._",
		"If you lose it, everything is lost and nobody can help you.\n",
	)

	return completeSetup(core, newName, newPhrase, newPin)
}

func askIfReady(feed slate) chan bool {
	return affirm(feed, "setup:ready?", "Are you ready?", "Ready!", "Not yet...")
}

func askIfReadyNow(feed slate) chan bool {
	return affirm(feed, "setup:ready?", "Okay, tell me when you're ready.",
		"Ready!", "Not yet...")
}

func proposeName(feed slate, name string) chan bool {
	return affirmSecret(feed, "setup:okName?", "Does this `session name` look okay to you?", name,
		"Yes, continue", "No, make another")
}

func proposeAnotherName(feed slate, name string) chan bool {
	return affirmSecret(feed, "setup:okName?", "How about this one?", name,
		"Yes, continue", "No, make another")
}

func proposePhrase(feed slate, phrase string) chan bool {
	return affirmSecret(feed, "setup:okPassphrase?", "Does this `passphrase` look okay to you?", phrase,
		"Yes, continue", "No, make another")
}

func proposeAnotherPhrase(feed slate, phrase string) chan bool {
	return affirmSecret(feed, "setup:okPassphrase?", "How about this one?", phrase,
		"Yes, continue", "No, make another")
}

func proposePin(feed slate, pin string) chan bool {
	return affirmSecret(feed, "setup:okPIN?", "Does this `PIN` look okay?", pin,
		"Yes, continue", "No, make another")
}

func promptWrittenDown(feed slate) chan bool {
	return affirm(feed, "setup:phraseWrittenDown?", "Tell me when you're finished writing...",
		"Ready!", "Hold on...")
}

func promptWrittenDownAgain(feed slate) chan bool {
	return affirm(feed, "setup:phraseWrittenDown?", "Okay just let me know when you're ready...",
		"Ready!", "Hold on...")
}
