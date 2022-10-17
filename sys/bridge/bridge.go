package bridge

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gorilla/websocket"
	logging "github.com/ipfs/go-log/v2"
	nanoid "github.com/matoous/go-nanoid/v2"
	"github.com/projectdiscovery/sslcert"

	"slater/core/msg"
)

var log = logging.Logger("slater:bridge")

func init() {
	logging.SetAllLoggers(logging.LevelDebug)
}

type Bridge struct {
	Input    chan any
	Output   chan any
	Sessions map[string]*websocket.Conn
	upgrader websocket.Upgrader
}

// TODO prob use some other identifier for sessions...

type InputSendMessage struct {
	Session string
	Message *msg.Message
}

type OutputSessionStart struct {
	Session string
}

type OutputSessionResume struct {
	Session string
}

type OutputSessionQuit struct {
	Session string
}

type OutputReceivedMessage struct {
	Session string
	Message *msg.Message
}

func Start() *Bridge {
	bridge := &Bridge{
		Input:    make(chan any),
		Output:   make(chan any),
		Sessions: make(map[string]*websocket.Conn),
		upgrader: websocket.Upgrader{
			// TODO Examine buffer sizes.
			// This is a quick guess aiming at mostly small messages,
			// and accepting a big increase in latency for larger messages like blobs.
			ReadBufferSize:  256,
			WriteBufferSize: 256,

			// TODO I'm not sure if compression is worth the overhead for local IPC,
			// but what about over the LAN? (docs/bridge.md)
			//EnableCompression: true,

			// TODO Check that connection is from localhost,
			// or a host in the whitelist. (docs/bridge.md)
			//CheckOrigin: func(r *http.Request) bool {
			//	return true
			//},
		},
	}

	http.HandleFunc("/", bridge.Session)

	go func() {
		tlsOptions := sslcert.DefaultOptions
		tlsOptions.Host = "localhost"
		tlsConfig, err := sslcert.NewTLSConfig(tlsOptions)
		if err != nil {
			log.Fatal(err)
		}

		server := &http.Server{
			TLSConfig: tlsConfig,
		}

		listener, err := net.Listen("tcp", ":0")
		if err != nil {
			panic(err)
		}

		port := listener.Addr().(*net.TCPAddr).Port
		fmt.Println(port)

		err = server.ServeTLS(listener, "", "")

		if err != nil {
			fmt.Println(err)
			log.Fatal("slater: ", err)
		}
	}()

	go func() {
		for {
			select {
			case input := <-bridge.Input:
				switch input.(type) {
				case InputSendMessage:
					mess := input.(InputSendMessage)

					sock, there := bridge.Sessions[mess.Session]
					if !there {
						continue
					}

					err := sock.WriteJSON(mess.Message)
					if err != nil {
						log.Debug(err)
					}
				}
			}
		}
	}()

	return bridge
}

func (bridge *Bridge) Session(w http.ResponseWriter, r *http.Request) {
	sock, err := bridge.upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Debugf("websocket:", err)
		return
	}

	defer sock.Close()

	var id string = ""

	for {
		mt, mess, err := sock.ReadMessage()

		if err != nil {
			log.Debug(err)
			break
		}

		switch mt {
		case websocket.TextMessage:
			var bm map[string]any

			err := json.Unmarshal(mess, &bm)
			if err != nil {
				log.Debugf("failed to unmarshal ui input!!")
				return
			}

			kindField, there := bm["kind"]
			if !there {
				log.Debugf("missing kind!")
				return
			}
			kind, ok := kindField.(string)
			if !ok {
				log.Debugf("bad kind value!")
				return
			}

			switch kind {
			case "begin":
				id, err = nanoid.New()
				if err != nil {
					log.Fatal("could not generate id!\n", err)
				}

				bridge.Sessions[id] = sock
				bridge.Output <- OutputSessionStart{id}

			case "resume":
				sessionField, there := bm["session"]
				if !there {
					log.Debug("missing session!")
					return
				}
				session, ok := sessionField.(string)
				if !ok {
					log.Debug("bad session value!")
					return
				}
				bridge.Sessions[session] = sock
				bridge.Output <- OutputSessionResume{session}

			default:
				if id != "" {
					messageField, there := bm["message"]
					if !there {
						log.Debug("missing message!")
						return
					}
					message, ok := messageField.(msg.Message)
					if !ok {
						log.Debug("bad message!")
						return
					}
					bridge.Output <- OutputReceivedMessage{id, &message}
				}
			}

		case websocket.BinaryMessage:
			log.Debugf("binary messages not implemented...")
		}
	}
}
