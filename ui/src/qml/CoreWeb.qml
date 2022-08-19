
import QtQuick
import QtQuick.Controls
import QtWebView

Element {
    inner: Button {
        text: title

        onClicked: function(){
            var browser = webview.createObject(slate, {
                focus: true,
                x: slate.x,
                y: slate.y,
                z: 3,
                width: slate.width,
                height: slate.height,
                url: body,
            })
        }

        Component {
            id: webview
            WebView {
                id: view
                httpUserAgent: "some other browser (probably not Gecko)"

                Keys.onEscapePressed: ()=> {
                    view.destroy()
                }
            }
        }
    }
}
