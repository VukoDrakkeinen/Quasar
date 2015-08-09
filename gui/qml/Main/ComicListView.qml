import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../utils.js" as U
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
				text: styleData.row != -1 ? table.model.qmlGet(styleData.row, styleData.column, "display") : ""
				color: styleData.row != -1 ? table.model.qmlGet(styleData.row, styleData.column, "foreground") : Qt.rgba(0, 0, 0, 1)
				elide: styleData.row != -1 ? styleData.elideMode : Text.ElideNone
			}
		}
	}

	TableViewColumn {
		title: "Title"
		width: 200
		delegate: infoDelegate
	}
	
	TableViewColumn {
		title: "Chapters"
		width: 70
		delegate: infoDelegate
	}
	
	TableViewColumn {
		title: "Read"
		width: 90
		delegate: infoDelegate
	}
	
	TableViewColumn {
		title: "Last Checked"
		width: 140
		delegate: infoDelegate
	}
	
	TableViewColumn {
		title: "Status"
		width: 200
		delegate: Item {
			SaneProgressBar {
				id: spb
				anchors.fill: parent
				value: styleData.row != -1 ? table.model.qmlGet(styleData.row, styleData.column, "progress") : 0
			}
			
			Label {
				visible: true
				anchors.left: parent.left
				anchors.right: parent.right
				anchors.verticalCenter: parent.verticalCenter
				anchors.margins: 4
				text: {
					var status = styleData.row != -1 ? table.model.qmlGet(styleData.row, styleData.column, "status") : UpdateStatus.Error
					if (status == UpdateStatus.NoUpdates) {	return qsTr("No Updates") }
					if (status == UpdateStatus.Updating) {
						spb.indeterminate = true //temporary?
						return qsTr("Updating...")
					}
					if (status == UpdateStatus.NewChapters) { return qsTr("New Chapters") }
					if (status == UpdateStatus.Error) { return qsTr("ERROR") }
					return "???"
				}
				color: styleData.row != -1 ? table.model.qmlGet(styleData.row, styleData.column, "foreground") : Qt.rgba(0, 0, 0, 1)
				elide: Text.ElideRight
			}
		}
	}
	
}