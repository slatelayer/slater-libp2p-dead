import QtQuick
import QtQml.Models
import QtQuick.Controls
import QtQuick.Layouts
import Qt.labs.qmlmodels

Item {
    id: slate

    anchors.fill: slates

    property string name
    property string username: 'user' // TODO temporary...

    property var backgroundURL: Style.dark ? background.dark : background.light

    property var background: ({
        dark: "qrc:///res/backgrounds/triangles-dark.png",
        light: "qrc:///res/backgrounds/triangles-light.png"
    })

    property var backgrounds: ({
        beanstalk: {
            light: "qrc:///res/backgrounds/beanstalk-light.png",
            dark: "qrc:///res/backgrounds/beanstalk-dark.png",
        },
        memphis: {
            light: "qrc:///res/backgrounds/memphis-mini-light.png",
            dark: "qrc:///res/backgrounds/memphis-mini-dark.png",
        },
        triangles: {
            light: "qrc:///res/backgrounds/triangles-light.png",
            dark: "qrc:///res/backgrounds/triangles-dark.png",
        },
        binding: {
            light: "qrc:///res/backgrounds/binding-light.png",
            dark: "qrc:///res/backgrounds/binding-dark.png",
        },
        natural: {
            light: "qrc:///res/backgrounds/natural-light.png",
            dark: "qrc:///res/backgrounds/natural-dark.png",
        },
        wall: {
            light: "qrc:///res/backgrounds/wall-light.png",
            dark: "qrc:///res/backgrounds/wall-dark.png",
        },
        luxury: {
            light: "qrc:///res/backgrounds/luxury-light.png",
            dark: "qrc:///res/backgrounds/luxury-dark.png",
        },
        papyrus: {
            light: "qrc:///res/backgrounds/papyrus-light.png",
            dark: "qrc:///res/backgrounds/papyrus-dark.png",
        },
        hex: {
            light: "qrc:///res/backgrounds/what-the-hex-light.png",
            dark: "qrc:///res/backgrounds/what-the-hex-dark.png",
        },
        maze: {
            light: "qrc:///res/backgrounds/maze-light.png",
            dark: "qrc:///res/backgrounds/maze-dark.png",
        },
        scales: {
            light: "qrc:///res/backgrounds/scales-light.png",
            dark: "qrc:///res/backgrounds/scales-dark.png",
        },
    })

    Image {
        anchors.fill: parent
        width: 400; height: 400
        source: backgroundURL
        fillMode: Image.Tile
        horizontalAlignment: Image.AlignLeft
        verticalAlignment: Image.AlignTop
    }

    property var bodies: ([])
    ListModel { id: model }

    property var elements: ({
        user: {
            messages: {},
            inputs: {},
        },
        core: {
            messages: {
                text: Qt.createComponent("CoreText.qml", Component.PreferSynchronous, slate),
                secretText: Qt.createComponent("CoreSecretText.qml", Component.PreferSynchronous, slate),
                web: Qt.createComponent("CoreWeb.qml", Component.PreferSynchronous, slate),
            },
            inputs: {
                text: Qt.createComponent("CoreTextInput.qml", Component.PreferSynchronous, slate),
                secretText: Qt.createComponent("CoreSecretTextInput.qml", Component.PreferSynchronous, slate),
                choice: Qt.createComponent("CoreChoiceInput.qml", Component.PreferSynchronous, slate),
            }
        },
    })

    DelegateChooser {
        id: chooser
        role: 'kind'

        DelegateChoice { roleValue: 'secretText'; delegate: elements.core.messages.secretText }
        DelegateChoice { roleValue: 'web'; delegate: elements.core.messages.web }
        DelegateChoice { /*roleValue: 'text';*/ delegate: elements.core.messages.text }
    }

    ColumnLayout {
        anchors.fill: parent
        spacing: 0
        opacity: 0.70

        Pane {
            id: headerPane

            Layout.fillWidth: true
            Material.elevation: 1
            z: 2

            Label {
                anchors.centerIn: parent
                verticalAlignment: Text.AlignVCenter
                textFormat: Text.MarkdownText
                text: `**${slate.name}**`
            }
        }

        ListView {
            id: elems
            model: model
            delegate: chooser

            Layout.fillWidth: true
            Layout.fillHeight: true
            Layout.margins: 20

            cacheBuffer: slate.height

            spacing: 20

            highlight: highlight
            highlightFollowsCurrentItem: true
            keyNavigationWraps: true
            displayMarginBeginning: headerPane.height
            displayMarginEnd: input.height
            
            property bool scrollDown: true
            //TODO disable scrollDown on scroll back, re-enable on return to bottom (maybe by button also TODO LOL)

            ScrollIndicator.vertical: ScrollIndicator {}
        }

        Pane {
            id: input
            
            Layout.fillWidth: true
            Layout.preferredHeight: 180
            padding: 0
            
            Material.elevation: 2

            focus: true
            activeFocusOnTab: true

            property var prompt
            property bool promptActive: !!prompt && inputs.currentIndex === inputs.idxs[prompt.kind]

            onPromptChanged: _ => {
                if (prompt) {
                    inputs.positionViewAtIndex(inputs.idxs[prompt.kind], PathView.Center)
                } else {
                    inputs.positionViewAtIndex(0, PathView.Center)
                }
            }

            PathView {
                id: inputs

                anchors.fill: parent
                clip: true
                
                snapMode: PathView.SnapToItem
                preferredHighlightBegin: 0.5
                preferredHighlightEnd: 0.5
                highlightMoveDuration: 420

                path: Path {
                    startX: input.width / 2
                    startY: -input.height
                    PathLine {
                        x: input.width / 2
                        y: input.height * 2
                    }
                }

                onCurrentItemChanged: _ => {
                    currentItem.forceActiveFocus()
                }

                Keys.onEscapePressed: _ => {
                    focus = false
                    elems.focus = true
                }

                Keys.onUpPressed: _ => {
                    decrementCurrentIndex()
                }
                Keys.onDownPressed: _ => {
                    incrementCurrentIndex()
                }

                // TODO Elements which only make sense in the context of their prompt,
                // such as `choice` should be inserted and removed from inputs as needed.

                property var idxs: ({
                    "text": 0,
                    "choice": 1, // TODO move...
                    "secretText": 2,
                })

                model: ListModel {
                    ListElement {kind: "text"}
                    ListElement {kind: "choice"} // TODO move
                    ListElement {kind: "secretText"}
                }
                
                delegate: DelegateChooser {
                    role: 'kind'
                    DelegateChoice { roleValue: 'text'; delegate: slate.elements.core.inputs.text }
                    DelegateChoice { roleValue: 'choice'; delegate: slate.elements.core.inputs.choice }
                    DelegateChoice { roleValue: 'secretText'; delegate: slate.elements.core.inputs.secretText }
                }
            }
        }
    }

    Component {
        id: highlight
        Item {
            height: 80
            y: elems.currentItem && elems.currentItem.y || -1
            Behavior on y {
                SpringAnimation {
                    spring: 3
                    damping: 0.2
                }
            }
        }
    }

    function appendMessage(msg) {
        console.log(JSON.stringify(msg))

        if (msg.event === "slate:clear") {
            model.clear()
            bodies = []
            return
        }

        if (msg.event === "slate:background:set") {
            var bg = backgrounds[msg.background]
            if (bg) {
                background = bg
            } else {
                Qt.callLater(_=>{
                    view.message({
                        slate: slate.name,
                        author: "system",
                        kind: "ui:error",
                        body: "no such background",
                        event: "ui:error:background:missing",
                    })
                })
            }
        }

        var alignment
        switch (msg.author) {
            case 'system':
                alignment = Qt.AlignLeft;
                break;
            case slate.username:
                alignment = Qt.AlignHCenter;
                break;
            default:
                alignment = Qt.AlignRight;
        }

        bodies.push(msg) // TODO manage the size of this when we add message list windowing

        // A new message will always "stomp on" a prompt that is already active.
        // So, basically: don't do that.
        // TODO make it impossible:
        // In the case of a script, it's pretty easy: the script code is going to
        // wait for a response from the last prompt it sent.
        // In the case of a human (maybe doing "wiz of oz"), the UI needs to block
        // sending another message. So send event messages for prompt activation...

        // ...actually, I'm going to remove the persistent input area,
        // and make input elements the same as output elements,
        // and classify all elements based on whether they are:
        // (TODO ontology) interactive, modal, etc
        // (now that I think I've finally settled this internal debate...)

        if (msg.prompt) {
            input.prompt = msg.prompt
        } else {
            input.prompt = null
        }

        var data = {
            kind: msg.kind,
            time: Qt.formatTime(new Date(msg.sent)),
            align: alignment,
            author: msg.author,
            title: msg.title || '',
            body: msg.body,
            prompt: !!msg.prompt,
        }

        model.append(data)

        // moved from elems.onCountChanged, because it stopped firing for all items (WTF)
        if (elems.scrollDown) {
            elems.currentIndex = elems.count - 1
        }
    }
}
