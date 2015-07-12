import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Options"

Window {
    title: qsTr("Comic Properties")
    flags: Qt.Dialog
    modality: Qt.WindowModal
    color: colorOf.window
    id: thisWindow
    width: mainLayout.implicitWidth + 2 * margin
    height: mainLayout.implicitHeight + 2 * margin
    minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
    minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin

    SystemPalette {
        id: colorOf
    }

    ColumnLayout {
        id: mainLayout
        anchors.fill: parent
        anchors.margins: 8

        GroupBox {
            Layout.fillWidth: true
            title: qsTr("Update notification mode:")
            NotificationModeChooser {
            }
        }

        GroupBox {
            Layout.fillWidth: true
            Layout.fillHeight: true
            title: qsTr("Sources:")
            GridLayout {
                anchors.fill: parent
                columns: 2

                Item {
                    width: buttoncol.width
                    height: buttoncol.height
                    Layout.alignment: Qt.AlignLeft | Qt.AlignTop
                    ColumnLayout {
                        id: buttoncol
                        Button {
                            text: qsTr("Add new source")
                            Layout.fillWidth: true
                        }
                        Button {
                            text: qsTr("Edit source")
                            Layout.fillWidth: true
                        }
                        Button {
                            text: qsTr("Remove source")
                            Layout.fillWidth: true
                        }
                        Button {
                            text: qsTr("Move up")
                            Layout.fillWidth: true
                        }
                        Button {
                            text: qsTr("Move down")
                            Layout.fillWidth: true
                        }
                    }
                }

                TableView {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    Layout.minimumWidth: 300

                    TableViewColumn {
                        role: "source"
                        title: qsTr("Source")
                        width: 100
                    }

                    TableViewColumn {
                        role: "priority"
                        title: qsTr("Priority")
                        width: 200
                    }
                }


                //Item { Layout.fillHeight: true; Layout.fillWidth: true }    //empty cell
                GridLayout {
                    Layout.fillHeight: true
                    Layout.fillWidth: true
                    Layout.columnSpan: 2
                    columns: 2

                    Label {
                        text: qsTr("Plugin:")
                    }
                    ComboBox {
                        Layout.fillWidth: true
                        model: [qsTr(
                                "[autodetect]"), "Batoto", "BakaUpdates"] //TODO
                    }

                    Label {
                        text: qsTr("URL:")
                    }
                    TextField {
                        Layout.fillWidth: true
                    }

                    Label {
                        id: markReadLabel
                        text: qsTr("Mark chapters as already read:")
                        MouseArea {
                            id: test
                            anchors.fill: parent
                            onReleased: {
                                markReadCheckBox.pressed = false
                                markReadCheckBox.checked = !markReadCheckBox.checked
                            }
                            onPressed: {
                                markReadCheckBox.pressed = true
                            }
                        }
                    }
                    CheckBox {
                        Layout.fillWidth: true
                        id: markReadCheckBox
                    }
                }
            }
        }

        OptionsBottomButtons {
        }
    }
}
