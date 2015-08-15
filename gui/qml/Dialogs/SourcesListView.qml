import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Options"
import "../Workarounds"

GridLayout {
	id: root
	implicitWidth: 300
	implicitHeight: 200
	columns: 2
	
	property var model: null
	
	QtObject {
		id: internal
		property var pluginNames: []
	}
	
	function populatePluginNames(pluginNames) {
		internal.pluginNames = pluginNames
	}
	
	function populatePluginComboBox(humanReadablePluginNames) {
		pluginChooser.model = humanReadablePluginNames
	}
	
	function pluginNameForIndex(index) {
		return internal.pluginNames[index]
	}
	
	function indexForPluginName(pluginName) {
		return internal.pluginNames.indexOf(pluginName)
	}
	
	Item {
		width: buttoncol.width
		height: buttoncol.height
		Layout.alignment: Qt.AlignLeft | Qt.AlignTop
		ColumnLayout {
			id: buttoncol
			Button {
				Layout.fillWidth: true
				text: qsTr("Add")
				action: Action {
					onTriggered: {
						var max = sources.model.count
						var url = urlField.text
						var sourceIdx = pluginChooser.currentIndex
						var pluginName = pluginChooser.currentText
						if (sourceIdx == 0) {
							pluginName = quasarCore.pluginAutodetect(url)
							sourceIdx = root.indexForPluginName(pluginName)
						}
						if (pluginName !== "") {
							sources.model.append({"priority": -1, "sourceIdx": sourceIdx, "url": url, "markAsRead": markReadCheckBox.checked})
							urlField.text = ""	//reset or not?
							pluginChooser.currentIndex = 0
							markReadCheckBox.checked = false
							sources.selection.clear()
							sources.selection.select(max)
							sources.currentRow = max
						} else {
							flashUrlRed.start()
						}
					}
				}
			}
			Button {
				Layout.fillWidth: true
				text: qsTr("Edit")
				action: Action {
					onTriggered: {
						var row = sources.currentRow
						if (row !== -1) {
							var data = sources.model.get(row)
							pluginChooser.currentIndex = data.sourceIdx
							urlField.text = data.url
							markReadCheckBox.checked = data.markAsRead
							sources.currentRow = -1
							sources.selection.clear()
							sources.model.remove(row, 1)
						}
					}
				}
			}
			Button {
				Layout.fillWidth: true
				text: qsTr("Remove")
				action: Action {
					onTriggered: {
						//var row = sources.currentRow	//currentRow may not be the first row in selection sometimes, so we can't use it :(
						var row = -1;
						sources.selection.forEach(function (i) {if (row == -1) row = i;})	//so wasteful ;_;
						var count = sources.selection.count
						var max = sources.model.count
						if (row !== -1) {
							sources.model.remove(row, count)
						}
						
						if (row+count === max) {
							sources.selection.clear()	//deselect nonexisting (well, anymore) items
							if (row != 0) {
								sources.selection.select(row-1)
							}
						}
					}
				}
			}
			Button {
				Layout.fillWidth: true
				text: qsTr("Move up")
				action: Action {
					onTriggered: {
						//var row = sources.currentRow	//currentRow may not be the first row in selection sometimes, so we can't use it :(
						var row = -1;
						sources.selection.forEach(function (i) {if (row == -1) row = i;})	//so wasteful ;_;
						if (row === 0) return;
						var count = sources.selection.count
						
						sources.model.move(row, row-1, count)
						sources.selection.clear()
						sources.selection.select(row-1, row+count-2)
					}
				}
			}
			Button {
				Layout.fillWidth: true
				text: qsTr("Move down")
				action: Action {
					onTriggered: {
						//var row = sources.currentRow	//currentRow may not be the first row in selection sometimes, so we can't use it :(
						var row = -1;
						sources.selection.forEach(function (i) {if (row == -1) row = i;})	//so wasteful ;_;
						var count = sources.selection.count
						if ((row+count) === sources.rowCount) return;
						sources.model.move(row, row+1, count)
						sources.selection.clear()
						sources.selection.select(row+1, row+count)
					}
				}
			}
			
		}
	}
	
	TableViewPatched {
		selectionMode: SelectionMode.ExtendedSelection
		Layout.fillHeight: true
		Layout.fillWidth: true
		Layout.minimumWidth: 300
		id: sources
		model: root.model
		
		TableViewColumn {
			role: "priority"
			title: qsTr("Priority")
			width: 50
			delegate: Item {
				Label {
					anchors.fill: parent
					anchors.leftMargin: 4
					text: styleData.row+1
				}
			}
		}
		
		TableViewColumn {
			role: "sourceIdx"
			title: qsTr("Source")
			width: 150
			delegate: Item {
				Label {
					anchors.fill: parent
					anchors.leftMargin: 4
					text: pluginChooser.model[styleData.value]
				}
			}
		}
		
		TableViewColumn {
			role: "url"
			title: qsTr("URL")
			width: 350
		}
		
		TableViewColumn {
			role: "markAsRead"
			title: qsTr("Mark")
			width: 40
			delegate: Item {
				Label {
					anchors.fill: parent
					anchors.leftMargin: 12
					text: (styleData.value == true ? "✔" : "❌")
				}
			}
		}
	}
	
	
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
			id: pluginChooser
		}
		
		Label {
			text: qsTr("URL:")
		}
		TextField {
			id: urlField
			Layout.fillWidth: true
			
			SequentialAnimation {
				id: flashUrlRed
				ColorAnimation { target: urlField; property: "textColor"; to: "red"; easing.type: Easing.OutQuad; duration: 400 }
				ColorAnimation { target: urlField; property: "textColor"; to: urlField.textColor; easing.type: Easing.InQuad; duration: 400 }
			}
			
		}
		
		Label {
			text: qsTr("Mark chapters as already read:")
			MouseArea {
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
