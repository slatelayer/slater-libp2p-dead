package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSetupAndDiscovery(t *testing.T) {
	var timeout time.Duration = 30

	testdir := "/tmp/slatertest"

	defer func() {
		os.RemoveAll(testdir)
	}()

	core1 := Start(filepath.Join(testdir, "one"))
	core2 := Start(filepath.Join(testdir, "two"))

	core1.Input <- InputUISessionStart{Session: "test1"}
	core2.Input <- InputUISessionStart{Session: "test2"}

	sessionName := ""
	passphrase := ""
	pin := ""
	finished := false

	send := func(c chan (any), thing any) {
		select {
		case c <- thing:
			return
		default:
			return
		}
	}

	for !finished {
		coreMsg := <-core1.Output

		switch coreMsg.(type) {
		case OutputUIMessage:
			msg := coreMsg.(OutputUIMessage)
			inner := msg.Message

			mi, there := inner["msg"]
			if !there {
				continue
			}

			m := mi.(message)

			switch m["kind"] {
			case "text", "secretText":
				bodyField, there := m["body"]
				if !there {
					continue
				}

				body := bodyField.(string)

				if strings.Contains(body, "new user") {
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "you ready") {
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "`session name`") {
					sessionName = m["secretText"].(string)
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "`passphrase`") {
					passphrase = m["secretText"].(string)
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "`PIN`") {
					pin = m["secretText"].(string)
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "finished") {
					prompt := m["prompt"].(message)
					event := prompt["event"]
					kind := prompt["kind"]
					send(core1.Input, InputUIMessage{Session: "test1", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(0), "event": event, "kind": kind, "slate": "setup"}}})
					continue
				}

				if strings.Contains(body, "Don't lose it") {
					finished = true // last message
					break
				}

				continue

			default:
				continue

			}
		default:
			continue
		}

	}

	for {
		select {
		case <-time.After(timeout * time.Second):
			t.Fatalf("failed to connect after %d seconds", timeout)

		case coreMsg := <-core2.Output:
			switch coreMsg.(type) {

			case OutputConnectedOtherDevice:
				msg := coreMsg.(OutputConnectedOtherDevice)
				t.Logf("connected to %s", msg.device)
				return

			case OutputUIMessage:
				msg := coreMsg.(OutputUIMessage)
				inner := msg.Message

				mi, there := inner["msg"]
				if !there {
					continue
				}

				m := mi.(message)

				switch m["kind"] {
				case "text":
					bodyField, there := m["body"]
					if !there {
						continue
					}

					body := bodyField.(string)

					if strings.Contains(body, "new user") {
						prompt := m["prompt"].(message)
						event := prompt["event"]
						kind := prompt["kind"]
						send(core2.Input, InputUIMessage{Session: "test2", Message: message{"kind": "msg", "msg": map[string]any{"choice": float64(1), "event": event, "kind": kind, "slate": "setup"}}})
						continue
					}

					if strings.Contains(body, "session") {
						prompt := m["prompt"].(message)
						event := prompt["event"]
						kind := prompt["kind"]
						send(core2.Input, InputUIMessage{Session: "test2", Message: message{"kind": "msg", "msg": map[string]any{"body": sessionName, "event": event, "kind": kind, "slate": "setup"}}})
						continue
					}

					if strings.Contains(body, "passphrase") {
						prompt := m["prompt"].(message)
						event := prompt["event"]
						kind := prompt["kind"]
						send(core2.Input, InputUIMessage{Session: "test2", Message: message{"kind": "msg", "msg": map[string]any{"secretText": passphrase, "event": event, "kind": kind, "slate": "setup"}}})
						continue
					}

					if strings.Contains(body, "PIN") {
						prompt := m["prompt"].(message)
						event := prompt["event"]
						kind := prompt["kind"]
						send(core2.Input, InputUIMessage{Session: "test2", Message: message{"kind": "msg", "msg": map[string]any{"secretText": pin, "event": event, "kind": kind, "slate": "setup"}}})
						continue
					}

					continue

				default:
					continue
				}

			default:
				continue
			}
		}
	}

}
