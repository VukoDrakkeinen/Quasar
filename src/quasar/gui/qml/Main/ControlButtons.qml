import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtQuick.Controls 1.4

///Rectangle {
Item {
	Layout.minimumWidth: implicitWidth
	id: root
	implicitWidth: 0
	implicitHeight: buttons.childrenRect.height
    ///color: colorOf.window
    ///SystemPalette { id: colorOf }
    
    function calculateWidth() {
		return Array.prototype.slice.apply(buttons.children)
		.map(
			function(btn) {
				return btn.implicitWidth;
			}
		)
		.reduce(
			function(prev, curr, idx) {
				return prev > curr ? prev : curr;
			}
		)
    }

    ColumnLayout {
		id: buttons
		anchors.top: parent.top
		anchors.left: parent.left
		anchors.right: parent.right

        Button {
            Layout.fillWidth: true
			text: qsTr("Add comic")
			action: Action {
				onTriggered: addComic.show()
			}
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Remove comic")
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Select all")
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Check for updates")
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Download")
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Read")
        }
        Button {
            Layout.fillWidth: true
            text: qsTr("Properties")
            action: Action {
                onTriggered: properties.show()
            }
        }
        Component.onCompleted: root.implicitWidth = calculateWidth()
    }
}
