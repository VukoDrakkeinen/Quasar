import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2

Window {
	title: qsTr("Options")
	flags: Qt.Dialog
	modality: Qt.WindowModal
	color: colorOf.window
	id: thisWindow
	width: 600
	height: 500
	minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
	minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin
	
	SystemPalette {
		id: colorOf
	}

	ColumnLayout {
		anchors.fill: parent
		anchors.margins: 8
		id: mainLayout

		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Default update notification mode:")
			NotificationModeChooser {}
		}

		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Downloads location:")
			RowLayout {
				anchors.fill: parent

				TextField {
					Layout.fillWidth: true
					text: "/home/vuko/Downloads"
				}
				Button {
					text: "IKONA"
				}
			}
		}

		GroupBox {
			Layout.fillWidth: true
			Layout.fillHeight: true
			title: qsTr("Plugins:")
			RowLayout {
				anchors.fill: parent
				Item {
					width: buttoncol.width
					height: buttoncol.height
					Layout.alignment: Qt.AlignLeft | Qt.AlignTop
					ColumnLayout {
						id: buttoncol
						Button {
							text: qsTr("Install plugin")
							Layout.fillWidth: true
						}
						Button {
							text: qsTr("Disable plugin")
							Layout.fillWidth: true
						}
						Button {
							text: qsTr("Remove plugin")
							Layout.fillWidth: true
						}
					}
				}

				TableView {
					Layout.fillHeight: true
					Layout.fillWidth: true
					Layout.minimumWidth: 300

					TableViewColumn {
						role: "name"
						title: qsTr("Plugin")
						width: 100
					}

					TableViewColumn {
						role: "status"
						title: qsTr("Status")
						width: 200
					}
				}
			}
		}

		OptionsBottomButtons {
		}
	}
}
