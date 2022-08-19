import QtQuick
import QtQuick.Controls

Item {

    property var components: ({
        text: debug//text
    })

    Component {
        id: debug

        Pane {
            property string author
            property string kind
            property int sent
            
            readonly property double maxWidth: 700
            readonly property double minWidth: parent.children[0].width
            readonly property alias textWidth: metrics.width
            readonly property bool wrap: textWidth > maxWidth

            Material.elevation: 2
            z: 2

            property alias text: label.text

            Label {
                id: label

                width: Math.min(Math.max(minWidth, textWidth), maxWidth)
                horizontalAlignment: wrap ? Text.AlignLeft : Text.AlignHCenter
                verticalAlignment: Text.AlignVCenter
                wrapMode: wrap ? Text.Wrap : Text.NoWrap
                clip: true

                //padding: 20

                font.pixelSize: 18
                textFormat: Text.MarkdownText
                color: Material.foreground

                /*
                background: Rectangle{
                    color: Material.primary
                    opacity: 0.25
                    border.color: Material.accent;
                    anchors.fill: parent
                }
                */

                TextMetrics {
                    id: metrics
                    font: label.font
                    text: label.text
                }

                Component.onCompleted: function () {
                    console.log(parent.children[0].width, minWidth)
                }
            }
        }
    }

    Component {
        id: text

        Item {
            id: item
            
            property string text
            // property color fgColor:parent.fgColor

            anchors.left: parent.left
            anchors.right: parent.right
            Text {
                text:item.text
                // color:parent.fgColor
                Component.onCompleted: function () {console.log(item.text)}
            }
        }
    }
}
