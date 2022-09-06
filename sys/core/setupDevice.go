package core

func setupDevice(core *Core, feed slate, say func(...string)) (datastore, *node) {
	say("Awesome, let's set up this device.")

	sessionName := <-promptSessionName(feed)
	passphrase := <-promptPassphrase(feed)
	pinNumber := <-promptPIN(feed)

	return completeSetup(core, sessionName, passphrase, pinNumber)
}

func promptSessionName(feed slate) chan string {
	return prompt(feed, "sessionName", "Punch in your `session name`")
}
