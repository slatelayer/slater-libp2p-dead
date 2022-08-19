import QtQuick
import QtQuick.Controls
import QtQuick.Layouts

InputElement {
	id: element

	property var prompt: input.prompt

	ListModel {id: choices}
	
	onPromptChanged: _=> {
		if (prompt && prompt.choices){
			prompt.choices.forEach((choice, idx)=>choices.append({choice,idx}))
			forceActiveFocus()
		} else {
			choices.clear()
		}
	}

	onFocusChanged: {
		if (focus) {
			rep.forceActiveFocus()
		}
	}

	inner: RowLayout {
		anchors.horizontalCenter: parent.horizontalCenter
		width: Math.min(input.width, 700)
		spacing: 33

		Repeater {
			id: rep
			model: choices

			Button {
				Layout.fillWidth: true
				Material.elevation: 2
				text: model.choice
				onClicked: _=>{
					var msg = {
	                    slate: slate.name,
	                    author: slate.username,
	                    kind: "text",
	                    event: input.prompt.event,
	                    body: model.choice,
	                    choice: model.idx,
	                }
	                view.message(msg)
				}
			}
		}
	}

	Keys.onEscapePressed: _ => {
        focus = false
        input.focus = true
    }

	Keys.onUpPressed: _ => {
        inputs.decrementCurrentIndex()
    }
    Keys.onDownPressed: _ => {
        inputs.incrementCurrentIndex()
    }
}
