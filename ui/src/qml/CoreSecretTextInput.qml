import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

InputElement {
    id: element

    onFocusChanged: {
        if (focus) {
            field.forceActiveFocus()
        }
    }

    inner: RowLayout {
        anchors.horizontalCenter: parent.horizontalCenter
        width: Math.min(input.width, 700)

        TextField {
            id: field
            placeholderText: "make sure nobody is looking"
            property bool hidden: true

            echoMode: hidden ? TextInput.Password : TextInput.Normal

            Layout.fillWidth: true
            Layout.alignment: Qt.AlignHCenter

            horizontalAlignment: TextInput.AlignHCenter
            verticalAlignment: TextInput.AlignVCenter

            width: Math.min(slate.width, 700)
            padding: 0
            leftPadding: button.width
            rightPadding: button.width

            font.pixelSize: 24

            Button {
                id: button
                anchors.left: parent.left
                flat: true
                text: field.hidden ? 'show' : 'hide'
                onClicked: _=> field.hidden = !field.hidden
            }
            
            onAccepted: _ => {
                if (text === "") return

                view.message({
                    slate: slate.name,
                    author: slate.username,
                    kind: "secretText",
                    body: '`' + input.prompt.event + '`',
                    secretText: text,
                    event: input.prompt.event,
                })
                text = ''
                elems.scrollDown = true // are we doing this in every input element? is it the best way? maybe do send out through `slate.message` slot first, and do it there?
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
}
