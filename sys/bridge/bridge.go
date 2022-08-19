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
	Message map[string]any
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
	Message map[string]any
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
					msg := input.(InputSendMessage)

					sock, there := bridge.Sessions[msg.Session]
					if !there {
						continue
					}

					err := sock.WriteJSON(msg.Message)
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
		mt, msg, err := sock.ReadMessage()

		if err != nil {
			log.Debug(err)
			break
		}

		switch mt {
		case websocket.TextMessage:
			var m map[string]any

			err := json.Unmarshal(msg, &m)
			if err != nil {
				log.Debugf("failed to unmarshal ui input!!")
				return
			}

			kind := m["kind"]

			switch kind {
			case "begin":
				id, err = nanoid.New()
				if err != nil {
					log.Fatal("could not generate id!\n", err)
				}

				bridge.Sessions[id] = sock
				bridge.Output <- OutputSessionStart{id}

			case "resume":
				id = m["session"].(string)
				bridge.Sessions[id] = sock
				bridge.Output <- OutputSessionResume{id}

			default:
				if id != "" {
					bridge.Output <- OutputReceivedMessage{id, m}
				}
			}

		case websocket.BinaryMessage:
			log.Debugf("binary messages not implemented...")
		}
	}
}
