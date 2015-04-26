import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1

RowLayout {
    Layout.fillWidth: true
    Button {
        Layout.alignment: Qt.AlignLeft
        text: qsTr("Defaults")
    }
    Item {
        Layout.fillWidth: true
    } //filler
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("OK")
    }
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("Apply")
    }
    Button {
        Layout.alignment: Qt.AlignRight
        text: qsTr("Cancel")
        action: Action {
            onTriggered: thisWindow.hide()
        }
    }
}
