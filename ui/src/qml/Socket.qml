import QtQuick 2.0
import QtWebSockets 1.0

Item {
    // TODO set this in c++ after launching the daemon (with a random port)
    property string daemonURL: "ws://localhost:9999"

    WebSocket {
        id: socket

        url: daemonURL

        //active: false

        onTextMessageReceived: {
            console.log(message)
        }

        onStatusChanged: if (socket.status == WebSocket.Error) {
                             console.log("Error: " + socket.errorString)
                         } else if (socket.status == WebSocket.Open) {
                             socket.sendTextMessage("Hello World")
                         } else if (socket.status == WebSocket.Closed) {
                             //messageBox.text += "\nSocket closed"
                             console.log("socket closed")
                         }
    }
}
