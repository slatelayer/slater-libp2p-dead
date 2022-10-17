package core

import (
	"slater/core/slate"
	"slater/core/store"
)

func setupDevice(core *Core, feed slate.Slate, say func(...string)) (store.Store, *node) {
	say("Awesome, let's set up this device.")

	sessionName := <-promptSessionName(feed)
	passphrase := <-promptPassphrase(feed)
	pinNumber := <-promptPIN(feed)

	return completeSetup(core, sessionName, passphrase, pinNumber)
}

func promptSessionName(feed slate.Slate) chan string {
	return prompt(feed, "sessionName", "Punch in your `session name`")
}
