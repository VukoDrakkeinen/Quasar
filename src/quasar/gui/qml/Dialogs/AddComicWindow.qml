import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Options"
import "../Workarounds"

Window {
	title: qsTr("Add Comic")
	flags: Qt.Dialog
	modality: Qt.WindowModal
	color: colorOf.window
	id: thisWindow
	width: mainLayout.implicitWidth + 2 * margin + 300
	height: mainLayout.implicitHeight + 2 * margin
	minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
	minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin
	
	SystemPalette {
		id: colorOf
	}
	
	property var pluginNames: []
	
	function resetAndShow() {
		sourcesModel.clear()
		var names = quasarCore.pluginNames()
		pluginNames = Array.prototype.concat("[autodetect]", names[0])
		pluginChooser.model = Array.prototype.concat(qsTr("[Autodetect]"), names[1])
		this.show()
	}
	
	ColumnLayout {
		id: mainLayout
		anchors.fill: parent
		anchors.margins: 8
		
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
							Layout.fillWidth: true
							text: qsTr("Add")
							action: Action {
								onTriggered: {
									var max = sourcesModel.count
									var url = urlField.text
									var pluginName = quasarCore.pluginAutodetect(url)
									if (pluginName !== "") {
										sourcesModel.append({"priority": -1, "sourceIdx": thisWindow.pluginNames.indexOf(pluginName), "url": url, "markAsRead": markReadCheckBox.checked})
										urlField.text = ""	//reset or not?
										pluginChooser.currentIndex = 0
										markReadCheckBox.checked = false
										sources.selection.clear()
										sources.selection.select(max)
										sources.currentRow = max
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
										var data = sourcesModel.get(row)
										pluginChooser.currentIndex = data.sourceIdx
										urlField.text = data.url
										markReadCheckBox.checked = data.markAsRead
										sourcesModel.remove(row, 1)
										sources.selection.clear()
										sources.currentRow = -1
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
									var max = sourcesModel.count
									if (row !== -1) {
										sourcesModel.remove(row, count)
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
									
									sourcesModel.move(row, row-1, count)
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
									sourcesModel.move(row, row+1, count)
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
					model: sourcesModel
					
					ListModel {
						id: sourcesModel
					}
					
					TableViewColumn {
						role: "priority"
						title: qsTr("Priority")
						width: 50
						delegate: Item {
						 	Label {
						 		anchors.verticalCenter: parent.verticalCenter
						 		text: " " + (styleData.row+1)	//TODO: remove space hack
						 	}
						 }
						 
					}
					
					TableViewColumn {
						role: "sourceIdx"
						title: qsTr("Source")
						width: 150
						delegate: Item {
							Label {
								anchors.verticalCenter: parent.verticalCenter
								text: " " + pluginChooser.model[styleData.value] //TODO: remove space hack
							}
						}
					}
					
					TableViewColumn {
						role: "url"
						title: qsTr("URL")
						width: 350
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
						id: pluginChooser
						Layout.fillWidth: true
					}
					
					Label {
						text: qsTr("URL:")
					}
					TextField {
						id: urlField
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
		
		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Update notification mode:")
			NotificationModeChooser {}
		}
		
		OptionsBottomButtons {}
	}
}
