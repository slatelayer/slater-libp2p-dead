
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Element {
    inner: ColumnLayout {

        Label {
            id: label
            text: body

            width: Math.min(Math.max(minWidth, shadow.width), maxWidth)
            wrapMode: Text.Wrap
            topPadding: 10

            font.pixelSize: 18
            textFormat: Text.MarkdownText

            Label {
                id: shadow
                visible: false
                text: body
                font.pixelSize: parent.font.pixelSize
                textFormat: Text.MarkdownText
            }
        }

        RowLayout {
            Button {
                flat: true
                text: secret.hidden ? 'show' : 'hide'
                onClicked: _=> secret.hidden = !secret.hidden
            }

            Label {
                id: secret
                property bool hidden: true
                property string plain: bodies[index] && bodies[index].secretText || ''
                property string obfuscate: plain.replace(/./g, "x")
                text: hidden ? obfuscate : plain

                width: hidden ? Math.max(minWidth, secretshadow.width) : Math.min(Math.max(minWidth, secretshadow.width), maxWidth)
                wrapMode: hidden ? Text.NoWrap : Text.Wrap
                elide: Text.ElideRight

                font.family: "Roboto Mono" // monospace so the obfuscate is exactly the same width
                font.pixelSize: 18
                textFormat: Text.MarkdownText

                Label {
                    id: secretshadow
                    visible: false
                    text: parent.plain
                    font.family: secret.font.family
                    font.pixelSize: parent.font.pixelSize
                    textFormat: Text.MarkdownText
                }
            }
        }
    }
}
