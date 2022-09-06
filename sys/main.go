package main

import (
	"os"
	"path/filepath"

	logging "github.com/ipfs/go-log/v2"

	Bridge "github.com/yeahcorey/slater/bridge"
	Core "github.com/yeahcorey/slater/core"
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
				msg := uiMsg.(Bridge.OutputSessionStart)
				core.Input <- Core.InputUISessionStart{Session: msg.Session}

			case Bridge.OutputSessionResume:
				msg := uiMsg.(Bridge.OutputSessionResume)
				core.Input <- Core.InputUISessionResume{Session: msg.Session}

			case Bridge.OutputReceivedMessage:
				msg := uiMsg.(Bridge.OutputReceivedMessage)
				core.Input <- Core.InputUIMessage{Session: msg.Session, Message: msg.Message}
			}

		case coreMsg := <-core.Output:
			switch coreMsg.(type) {
			case Core.OutputUIMessage:
				msg := coreMsg.(Core.OutputUIMessage)
				bridge.Input <- Bridge.InputSendMessage{Session: msg.Session, Message: msg.Message}
			}
		}
	}
}
