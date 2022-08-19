import QtQuick
import QtQuick.Layouts
import QtQuick.Controls

Item {
    id: element
    default property alias inner: inside.data
    readonly property double minWidth: authorstamp.width + timestamp.width + (pane.padding * 2)
    readonly property double maxWidth: 700
    readonly property bool byUser: author === slate.username
    readonly property bool current: ListView.isCurrentItem
    readonly property bool prompt: model.prompt
    readonly property real fadeDuration: 166
    readonly property real scaleDuration: 166

    width: elems.width
    height: pane.height
    
    opacity: 0
    PropertyAnimation {
        target: element;
        property: "opacity"; to: 1;
        duration: fadeDuration;
        running: true
    }

    Pane {
        id: pane
        anchors.top: parent.top
        anchors.left: align === Qt.AlignLeft ? parent.left : undefined
        anchors.right: align === Qt.AlignRight ? parent.right : undefined
        anchors.horizontalCenter: align === Qt.AlignHCenter ? parent.horizontalCenter : undefined
        contentWidth: inside.children[0].width
        contentHeight: inside.children[0].height + authorstamp.height

        Material.theme: Style.theme
        
        // animating the drop shadow looks like shit,
        // but the material style is temporary;
        // when we create the actual style, then we'll make nicely animated active / inactive states
        // and maybe the scale transform isn't the best way to do it anyway...

        /*
        Material.elevation: current ? 2 : 1
        
        Behavior on Material.elevation {
            PropertyAnimation {duration:scaleDuration}
        }

        transform: Scale {
            id: transform
            xScale: current ? 1.125 : 1.0
            yScale: current ? 1.125 : 1.0
            Behavior on xScale { PropertyAnimation { duration: scaleDuration } }
            Behavior on yScale { PropertyAnimation { duration: scaleDuration } }
        }
        */

        Material.elevation: 1
        
        padding: 15

        Label {
            id: authorstamp
            anchors.top: parent.top
            anchors.left: parent.left
            font.pixelSize: 12
            opacity: 0.6
            textFormat: Text.MarkdownText
            text: byUser ? '' : `**${author}**`
        }

        Label {
            id: timestamp
            anchors.top: parent.top
            anchors.left: authorstamp.right
            anchors.leftMargin: byUser ? 0 : 5
            font.pixelSize: 12
            opacity: 0.4
            textFormat: Text.MarkdownText
            text: `${time}`
        }

        Item {
            id: inside
            anchors.left: parent.left
            anchors.top: authorstamp.bottom
        }
    }
}
