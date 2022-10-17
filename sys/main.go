package main

import (
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log/v2"

	Bridge "slater/bridge"
	Core "slater/core"
)

var log = logging.Logger("slater")

func init() {
	logging.SetAllLoggers(logging.LevelDebug)
}

func main() {
	var rootPath string

	if len(os.Args) > 1 {
		alt := os.Args[1] // alternate root, to run 2 instances for testing
		if alt != "" {
			rootPath = alt
		}
	} else {
		home, _ := os.UserHomeDir()
		rootPath = filepath.Join(home, ".slater")
	}

	bridge := Bridge.Start()
	core := Core.Start(rootPath)

	for {
		select {
		case uiMsg := <-bridge.Output:
			switch uiMsg.(type) {
			case Bridge.OutputSessionStart:
				m := uiMsg.(Bridge.OutputSessionStart)
				core.Input <- Core.InputUISessionStart{Session: m.Session}

			case Bridge.OutputSessionResume:
				m := uiMsg.(Bridge.OutputSessionResume)
				core.Input <- Core.InputUISessionResume{Session: m.Session}

			case Bridge.OutputReceivedMessage:
				m := uiMsg.(Bridge.OutputReceivedMessage)
				core.Input <- Core.InputUIMessage{Session: m.Session, Message: m.Message}
			}

		case coreMsg := <-core.Output:
			switch coreMsg.(type) {
			case Core.OutputUIMessage:
				m := coreMsg.(Core.OutputUIMessage)
				bridge.Input <- Bridge.InputSendMessage{Session: m.Session, Message: m.Message}
			}
		}
	}
}
