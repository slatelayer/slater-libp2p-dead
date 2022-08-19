

# The Bridge


This implementation of Slater uses a websocket bridge between the core and the interface.

I considered using QT Remote Objects, and also Mangos / Nanomsg, but decided to use websockets so that I could quickly build an interface in QML since it has a websocket binding already provided. It also opens up the option of connecting easily with other UI implementations, and connecting from inside Slater's webview as well.

The interface program runs the Slater core with a random port argument, waits a sec, then tries to connect. If it fails, it will try again (and again...) with another random port.

Messages are passed using JSON (again, because QML has it already).

The bridge might also be used to connect other devices:
	1  before initial sync completes, so that a new device can work immediately
	2  to enable a thin / ephemeral client
	3  to control one device (or many) from another's input devices
