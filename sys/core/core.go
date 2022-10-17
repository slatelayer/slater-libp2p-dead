package core

import (
	"os"

	logging "github.com/ipfs/go-log/v2"

	"slater/core/msg"
	"slater/core/slate"
	"slater/core/store"
)

var log = logging.Logger("slater:core")

type Core struct {
	root     string
	store    store.Store
	host     *node
	devices  []string
	sessions map[string]session
	Input    chan any
	Output   chan any
}

func Start(rootPath string) Core {
	logging.SetLogLevel("slater:core", "debug")

	if err := os.MkdirAll(rootPath, 0700); err != nil {
		log.Panic(err)
	}

	core := Core{
		root:     rootPath,
		devices:  make([]string, 0),
		sessions: make(map[string]session),
		Input:    make(chan any, 128),
		Output:   make(chan any, 128),
	}

	go core.Run()

	return core
}

func (core *Core) Run() {
	for {
		select {
		case input := <-core.Input:
			switch input.(type) {
			case InputUISessionStart:
				msg := input.(InputUISessionStart)
				session := msg.Session
				go core.connect(session)

			case InputUISessionResume:
				msg := input.(InputUISessionResume)
				session := msg.Session
				go core.resumeSession(session)

			case InputUIMessage:
				msg := input.(InputUIMessage)
				session := msg.Session
				message := msg.Message
				go core.handleUIMessage(session, message)
			}
		}
	}
}

type InputUISessionStart struct {
	Session string
}

type InputUISessionResume struct {
	Session string
}

type InputUIMessage struct {
	Session string
	Message *msg.Message
}

type OutputUIMessage struct {
	Session string
	Message *msg.Message
}

type OutputConnectedOtherDevice struct {
	device string
}

func (core *Core) connect(sid string) {
	session := newSession(sid)
	core.sessions[sid] = session

	core.sendSessionID(sid)

	setup := "setup"
	core.sendAddSlate(sid, setup)

	feed := session.view.slates[setup]

	feed.On(slate.ALL, func(m *msg.Message) {
		core.sendMessage(sid, m)
	})

	store, host := runSetup(core, feed)
	core.store = store
	core.host = host

	core.handleNet()
}

func (core *Core) resumeSession(sid string) {
	session, there := core.sessions[sid]

	if !there {
		core.connect(sid)
		return
	}

	setup := "setup"
	core.sendAddSlate(sid, setup)

	s := session.view.slates[setup]

	// send everything on slate setup TODO PAGINATION!
	msgs, err := s.GetRange(0, -1)
	if err != nil {
		log.Debug(err)
	}
	for _, msg := range msgs {
		core.sendMessage(sid, msg)
	}
}

func (core *Core) handleUIMessage(sid string, m *msg.Message) {
	//log.Debugf("message from session %v:\n%v", sid, m)

	session, there := core.sessions[sid]

	if !there {
		log.Debug("discarded uiMessage from uninitialized session!")
		return
	}

	switch m.Kind {
	case "msg":
		content := m.Content

		slateField, there := content["slate"]
		if !there {
			log.Debug("missing slate field")
			return
		}
		slateName, ok := slateField.(string)
		if !ok {
			log.Debug("bad slate field")
			return
		}

		slate, there := session.view.slates[slateName]
		if there {
			slate.Write(m)
		} else {
			log.Debugf("failed write to missing slate %s", slateName)
		}
	}
}

func (core *Core) handleNet() {
	for {
		m := <-core.host.output
		content := m.Content

		slateField, there := content["slate"]
		if !there {
			log.Debug("missing slate field")
			continue
		}
		slateName, ok := slateField.(string)
		if !ok {
			log.Debug("bad slate field")
			continue
		}

		if m.Kind == "signet" {
			core.host.send(core.host.discoveryKey, &msg.Message{
				Slate: "setup",
				Kind:  "signet",
				Content: map[string]any{
					"signet": core.host.signet,
				},
			})

			core.Output <- OutputConnectedOtherDevice{device: m.Device}
		}

		for _, session := range core.sessions {
			sl8, there := session.view.slates[slateName]
			if there {
				sl8.Write(m)
				break
			} else {
				continue
			}
		}
	}
}

func (core *Core) sendMessage(sid string, m *msg.Message) {
	core.Output <- OutputUIMessage{sid, m}
}

func (core *Core) sendSessionID(sid string) {
	m := msg.Message{Kind: "session", Content: map[string]any{"session": sid}}
	core.Output <- OutputUIMessage{sid, &m}
}

func (core *Core) sendAddSlate(sid string, slate string) {
	m := msg.Message{Kind: "slate", Content: map[string]any{"slate": slate}}
	core.Output <- OutputUIMessage{sid, &m}
}

func (core *Core) sendPage(sid string, page []*msg.Message) {
	m := msg.Message{Kind: "page", Content: map[string]any{"page": page}}
	core.Output <- OutputUIMessage{sid, &m}
}
