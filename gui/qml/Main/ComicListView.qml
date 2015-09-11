import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QuasarGUI 1.0

TableView {
	id: table
	implicitWidth: 500
	selectionMode: SelectionMode.ExtendedSelection
	model: updateModel
	//itemDelegate: //can't use, bugged x_x
	
	Component {
		id: infoDelegate
		
		Item {
			Label {
				anchors.fill: parent
				anchors.leftMargin: 4
				text: styleData.value
				color: model ? model.foreground : styleData.textColor
				elide: styleData.row != -1 ? styleData.elideMode : Text.ElideNone
			}
		}
	}

	TableViewColumn {
		role: "title"
		title: "Title"
		width: 200
		delegate: infoDelegate
	}
	
	TableViewColumn {
		role: "chapters"
		title: "Chapters"
		width: 70
		delegate: infoDelegate
	}
	
	TableViewColumn {
		role: "read"
		title: "Read"
		width: 90
		delegate: infoDelegate
	}
	
	TableViewColumn {
		role: "time"
		title: "Last Checked"
		width: 140
		delegate: infoDelegate
	}
	
	TableViewColumn {
		role: "status"
		title: "Status"
		width: 200
		delegate: Item {
			SaneProgressBar {
				id: spb
				anchors.fill: parent
				visible: false
				value: model ? model.progress : 0
			}
			
			Label {
				anchors.left: parent.left
				anchors.right: parent.right
				anchors.verticalCenter: parent.verticalCenter
				anchors.margins: 4
				text: {
					var status = styleData.value
					spb.visible = false
					switch (status|0) {
						case UpdateStatus.NoUpdates: return qsTr("No Updates")
						case UpdateStatus.Updating: 
							spb.indeterminate = true //temporary?
							spb.visible = true
							return qsTr("Updating...")
						case UpdateStatus.NewChapters: return qsTr("New Chapters")
						case UpdateStatus.Error: return qsTr("ERROR")
						default: return "???"
					}
				}
				color: model ? model.foreground : styleData.textColor
				elide: Text.ElideRight
			}
		}
	}	
}