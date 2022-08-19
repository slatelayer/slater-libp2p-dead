import QtQuick
import QtQml.Models
import QtQuick.Controls

Pane {
    signal message (msg: variant)

    anchors.fill: parent
    padding: 0

    ObjectModel { id: model }

    ListView {
        id: slates
        model: model

        orientation: ListView.Horizontal
        snapMode: ListView.SnapToItem
        anchors.fill: parent
    }

    function addSlate (id: string) {
        var src = `Slate{id:${id}; name:"${id}"}`
        model.append(Qt.createQmlObject(src, slates, "slate"))
    }

    /*
    function sendMessage (msg: variant) {
        message (msg)
    }
    */

    function appendMessage (msg: variant) {
        var slate = find(msg.slate)

        if (!slate){
            return console.log("no such slate!")
        }

        slate.appendMessage(msg)
    }

    /*
    function insertMessage (slate, idx, msg) {
        var slate = find(msg.slate)

        if (!slate){
            return console.log("no such slate!")
        }

        slate.appendMessage(idx, msg)
    }
    */

    function addPage (slate: string, page) {
        console.log("page:\n"+ JSON.stringify(page))
    }

    function addElement (msg: variant) {
        var slate = find(msg.slate)
        slate.addElement(msg.name, msg.version, msg.src)
    }

    function find (slate: string) {
        // TODO maybe maintain an index? maybe only when the count exceeds some threshold?
        // ACTUALLY... since we are maintaing the view state on the core side, we will use an int idx...
        for (var i=0; i<model.count; ++i) {
            var it = model.get(i)
            if (it.name === slate) return it
        }
    }
}
