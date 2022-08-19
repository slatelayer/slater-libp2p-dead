
import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

Item {
	id: element

	default property alias inner: inside.data

	property real maxWidth: 700

	width: Math.min(inside.children[0].width, maxWidth)
    height: input.height

	opacity: Math.max(0, 1.0 - Math.abs(displacement) / 0.5)

    property real displacement: calculateDisplacement()

    function calculateDisplacement () {
        var view = element.PathView.view
        var displacement = view.count - index - view.offset
        var halfVisibleItems = 0.5 + (1 < view.count ? 1 : 0)
        if (displacement > halfVisibleItems)
            displacement -= view.count
        else if (displacement < -halfVisibleItems)
            displacement += view.count
        return displacement
    }

    Label {
        id: label
        font.pixelSize: 16
        //color: Material.accent
        opacity: 0.6
        textFormat: Text.MarkdownText
        text: input.promptActive &&  `\`${input.prompt.event}\`` || ''
        anchors.topMargin: 10
    }

    Item {
        id: inside
        anchors.centerIn: parent
    }
}
