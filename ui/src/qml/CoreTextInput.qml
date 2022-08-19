import QtQuick
import QtQuick.Controls

InputElement {
    id: element

    onFocusChanged: {
        if (focus) {
            tf.forceActiveFocus()
        }
    }

    inner: TextField {
        id: tf
        placeholderText: "message or /command"

        horizontalAlignment: TextInput.AlignHCenter
        verticalAlignment: TextInput.AlignVCenter

        width: Math.min(slate.width, 700)
        padding: 0

        anchors.centerIn: parent
        font.pixelSize: 24
        
        onAccepted: _ => {
            if (text === "") return

            // TODO BETTER COMMAND HANDLING
            // commands are going to the core soon, so chill (don't even look at this shit)

            var msg
            if (text.startsWith("//")) {
                var url = text.slice(2) // TODO parse & validate
                msg = {
                    slate: slate.name,
                    author: slate.username,
                    kind: "web",
                    body: `https://${url}`, // // DUH maybe support plain http? (`://` as protocol didn't work)
                    title: url,
                }
            } else if (text.startsWith("/cle")) {
                msg = {
                    slate: slate.name,
                    author: slate.username,
                    kind: "text",
                    body: "`clear`",
                    event: "slate:clear",
                }
            } else if (text.startsWith("/bg")) {
                var bg = text.split(' ')[1]
                msg = {
                    slate: slate.name,
                    author: slate.username,
                    kind: "text",
                    body: `set background to ${bg || 'default'}`,
                    event: "slate:background:set",
                    background: bg,
                }
            } else if (input.prompt) {
                msg = {
                    slate: slate.name,
                    author: slate.username,
                    kind: "text",
                    event: input.prompt.event,
                    body: text,
                }
            } else {
                msg = {
                    slate: slate.name,
                    author: slate.username,
                    kind: "text",
                    body: text,
                }
            }
            view.message(msg)
            text = ''
            elems.scrollDown = true
        }

        Keys.onEscapePressed: _ => {
            focus = false
            input.focus = true
        }

        Keys.onUpPressed: _ => {
            // TODO history
        }
        Keys.onDownPressed: _ => {
            // 
        }
    }
}