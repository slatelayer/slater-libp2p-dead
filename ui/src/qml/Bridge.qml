import QtQuick
import QtWebSockets
import Qt.labs.settings

Item {
    property bool connected: socket.status == WebSocket.Open
    property string dir: Qt.application.arguments[1] || ""
    
    signal message(variant msg)
    signal timeout()

    Timer {
        id: timer
        interval: 500
        onTriggered: timeout()
    }

    Settings {
        property alias port: socket.port
    }

    Connections {
        target: _core
        
        function onStart(){
            console.log("core started")
        }

        function onReady (port) {
            console.log("core running on " + port)
            socket.port = port
            socket.active = true
        }

        function onEnd (code) {
            console.log("core exited with " + code)
        }

        function onError (err) {
            console.log("core process error: " + err)
        }
    }

    WebSocket {
        id: socket

        property int port: -1

        url: "wss://localhost:" + port

        active: false

        onTextMessageReceived: function (bytes) {
            try {
                var msg = JSON.parse(bytes)
                message (msg)
            } catch (err) {
                console.log(err)
            }
        }
        
        onStatusChanged: function () {
            if (socket.status == WebSocket.Connecting) {
                console.log("bridge: connecting")
                timer.restart()

            } else if (socket.status == WebSocket.Error) {
                console.log(`bridge: ${socket.errorString}`)
                timer.stop()
                runCore()
                
            } else if (socket.status == WebSocket.Open) {
                console.log("bridge: connected")
                timer.stop()
                
            } else if (socket.status == WebSocket.Closed) {
                console.log("bridge: closed")
                
                if (port > 0)
                    socket.active = true // TODO correct reconnect?
            }
        }
    }

    Component.onCompleted: function(){
        if (socket.port == -1 || dir != "") {
            runCore()
        } else {
            socket.active = true
        }
    }

    function runCore() {
        console.log("runCore", dir)
        _core.run(dir)
    }

    function sendMessage (msg) {
        socket.sendTextMessage(msg)
    }
}
