
import QtQuick
import QtQuick.Controls

Element {
    inner: Label {
        id: label
        text: body

        readonly property alias textWidth: shadow.width
        readonly property bool wrap: textWidth > width

        width: Math.min(Math.max(minWidth, textWidth), maxWidth)
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
}
