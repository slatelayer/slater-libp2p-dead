package main

import (
	logging "github.com/ipfs/go-log/v2"

	Bridge "gitlab.com/slatersys/slater/bridge"
	Core "gitlab.com/slatersys/slater/core"
)

var log = logging.Logger("slater")

func init() {
	logging.SetAllLoggers(logging.LevelDebug)
}

func main() {
	bridge := Bridge.Start()
	core := Core.Start()

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
