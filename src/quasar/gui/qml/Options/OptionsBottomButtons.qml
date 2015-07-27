import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2

RowLayout {
    Layout.fillWidth: true
    
    signal defaults()
	signal oK()
	signal apply()
	signal cancel()
    
    Button {
        Layout.alignment: Qt.AlignLeft
        text: qsTr("Defaults")
		action: Action {
			onTriggered: defaults()
		}
    }
    Item {
        Layout.fillWidth: true
    } //filler
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("OK")
		action: Action {
			onTriggered: oK()
		}
    }
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("Apply")
		action: Action {
			onTriggered: apply()
		}
    }
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("Cancel")
        action: Action {
            onTriggered: cancel()
        }
    }
}
