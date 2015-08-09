import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QtQuick.Dialogs 1.2
import "../Options"
import "../Workarounds"

Window {
	title: qsTr("Remove comic(s)")
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
	
	QtObject {
		id: internal
		property var comicIndices: []
	}
	
	function resetAndShow(comicIndices, comicNames) {
		internal.comicIndices = comicIndices
		removableComics.model = comicNames
		deleteData.checked = false
		this.show()
	}
	
	ColumnLayout {
		id: mainLayout
		anchors.fill: parent
		anchors.margins: 8
		
		RowLayout {
			Layout.fillWidth: true
			Image {
				source: ""
			}
			
			Label {
				text: qsTr("Are you sure you want to remove the following comic(s)?")
			}
		}
		
		CheckBox {
			id: deleteData
			text: qsTr("Also delete downloaded data?")
		}
		
		GroupBox {
			Layout.fillWidth: true
			title: "Comics"
			
			Column {
				Repeater {
					id: removableComics
					model: []
					
					Label {
						text: "• " + modelData
					}
				}
			}
		}
	
		RowLayout {
			Layout.fillWidth: true
			
			Item {
				Layout.fillWidth: true
			} //padding
			
			Button {
				Layout.alignment: Qt.AlignRight
				text: qsTr("Remove")
				action: Action {
					onTriggered: {
						if (deleteData.checked) {
							console.log("quasarCore.deleteData(internal.comicIndices)")
						}
						console.log("quasarCore.removeComics(internal.comicIndices)")
						thisWindow.hide()
					}
				}
			}
			
			Button {
				Layout.alignment: Qt.AlignRight
				text: qsTr("Cancel")
				action: Action {
					onTriggered: {
						thisWindow.hide()
					}
				}
			}
		}
		
	}
}
