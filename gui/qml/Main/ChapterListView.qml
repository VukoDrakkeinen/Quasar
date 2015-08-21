import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QtQml.Models 2.2
import "../utils.js" as U

TreeView {
	id: tree
	implicitWidth: 500
	model: chapterModel
	selection: ItemSelectionModel {
		model: tree.model
	}
	selectionMode: SelectionMode.ExtendedSelection
	itemDelegate: Item {  
		Label {  
			anchors.verticalCenter: parent.verticalCenter
			text: styleData.value
			color: model ? model.foreground : styleData.textColor
			elide: styleData.elideMode
		}  
	}
	
	TableViewColumn {
		role: "display"
		title: "#"
		width: 70
	}
	
	TableViewColumn {
		role: "title"
		title: "Title"
		width: 350		
	}
	
	TableViewColumn {
		role: "scanlators"
		title: "Scanlators"
		width: 300
	}
	
	TableViewColumn {
		role: "lang"
		title: "Language"
		width: 100
	}
	
	TableViewColumn {
		role: "plugin"
		title: "Plugin"
		width: 100
	}
	
} 
