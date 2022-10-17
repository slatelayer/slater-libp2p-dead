import QtQuick
import QtQuick.Controls
import Qt.labs.settings
import QtQuick.Controls.Material

ApplicationWindow {
    id: window
    title: qsTr("slater")
    //visibility: Window.FullScreen
    visible: true
    opacity: 0

    width: 720
    height: 720

    PropertyAnimation {
        id: fadeIn;
        target: window;
        property: "opacity"; to: 1;
        duration: 500;
        running: true
    }

    Settings {
        id: settings
        property int port: 9999
        property string session: ""
        property bool pinLock: false
    }

    Material.theme:   Style.theme
    Material.primary: Material.BlueGrey
    Material.accent: Material.Grey
    
    View {
        id: view

        onMessage: function (msg) {
            // TODO actually pass this kinda stuff over the bridge also...
            // for example, in this case we actually want to set the mode on all devices...
            switch (msg.kind) {
                case "text": {
                    switch (msg.body) {
                        case "/dark":
                            return Style.setDarkMode(true)
                        case "/light":
                            return Style.setDarkMode(false)
                    }
                }
            }

            var msg = JSON.stringify({
                Kind: msg.kind,
                Message: msg,
            })
            bridge.sendMessage(msg)
        }
    }

    Bridge {
        id: bridge

        onTimeout: console.log("TIMEOUT!!")

        onMessage: function (msg) {
            var kind = msg.kind

            switch (kind) {
            case "session":
                settings.session = msg.session
                return

            case "msg":
                return view.appendMessage(msg.msg)

            case "page":
                return view.addPage(msg.slate, msg.page)

            case "slate":
                return view.addSlate(msg.slate)

            //case "element":
              //  return view.addElement(msg)

            default:
                console.log("unrecognized message: " + msg.kind)
            }
        }

        onConnectedChanged:
            if (bridge.connected) {
                if (settings.session && !Qt.application.arguments[1]) {
                    sendResume()
                } else {
                    sendBegin()
                }
            }

        function sendBegin () {
            var msg = JSON.stringify({
                kind: "begin"
            })
            bridge.sendMessage(msg)
        }

        function sendResume () {
            var msg = JSON.stringify({
                kind: "resume",
                session: settings.session
            })
            bridge.sendMessage(msg)
        }
    }
}
