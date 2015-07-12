import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Controls.Styles 1.4
import QtQuick.Window 2.2
import QtQuick.Layouts 1.2
import "Main"
import "Options"
import "Dialogs"
import "getProperties.js" as Utils

ApplicationWindow {
	title: "Quasar"
	property int margin: 11
	width: mainLayout.implicitWidth + 2 * margin
	height: mainLayout.implicitHeight + 2 * margin
	minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
	minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin

	OptionsWindow {
		id: options
	}

	PropertiesWindow {
		id: properties
	}
	
	AddComicWindow {
		id: addComic
	}

	menuBar: MenuBar {
		Menu {
			title: "Quasar"
			MenuItem {
				text: qsTr("Options")
				onTriggered: options.show()
			}
			MenuItem {
				text: qsTr("Quit")
				onTriggered: Qt.quit()
			}
		}
	}

	SplitView {
		id: mainLayout
		anchors.fill: parent
		orientation: Qt.Horizontal
		//anchors.margins: margin
		ControlButtons {}

		ColumnLayout {
		Layout.fillWidth: true
		
			TreeView {
				Layout.fillHeight: true
				Layout.fillWidth: true
				implicitWidth: 500
				model: chapterModel
				//model: 600
				
				TableViewColumn {
					role: "display"
					title: "#"
					width: 70
				}
				
				TableViewColumn {
					role: "title"
					title: "Title"
					width: 350
					/*delegate: Item {
						Label {
							anchors.verticalCenter: parent.verticalCenter
							text: chapterModel.qmlGet(styleData.row, styleData.column, "display")
							color: chapterModel.qmlGet(styleData.row, styleData.column, "foreground")
							elide: styleData.elideMode
						}
					}//*/
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
				
			}//*/
			/*TableView {
				Layout.fillHeight: true
				Layout.fillWidth: true
				implicitWidth: 500
				selectionMode: 2 //SelectionMode.ExtendedSelection
				model: updateModel
				itemDelegate: Label{
					anchors.fill:parent
					text: updateModel.qmlGet(styleData.row, styleData.column, "display")
					color: updateModel.qmlGet(styleData.row, styleData.column, "foreground")
					elide: styleData.elideMode
				}

				TableViewColumn {
					title: "Title"
					width: 200
				}

				TableViewColumn {
					title: "Chapters"
					width: 70
				}

				TableViewColumn {
					title: "Read"
					width: 90
				}

				TableViewColumn {
					title: "Last Updated"
					width: 140
				}

				TableViewColumn {
					delegate: Item {
											anchors.fill: parent

											ProgressBar {
												anchors.fill: parent
												maximumValue: 100
												value: {
													var status = updateModel.qmlGet(styleData.row, styleData.column, "status")
													if (status == 0) return 0
													if (status == 2) return 0
													if (status == 3) return 0
													return updateModel.qmlGet(styleData.row, styleData.column, "progress")
												}
											}
											Label {
												anchors.left: parent.left
												anchors.right: parent.right
												anchors.verticalCenter: parent.verticalCenter
												anchors.margins: 4
												text: {
													var status = updateModel.qmlGet(styleData.row, styleData.column, "status")
													if (status == 0) return "No Updates"
													if (status == 1) return "Updating..."
													if (status == 2) return "New Chapters"
													if (status == 3) return "ERROR"
													return "???"
												}
												color: updateModel.qmlGet(styleData.row, styleData.column, "foreground")
												elide: Text.ElideRight
											}
										}
					title: "Status"
					width: 200
				}
			}//*/

			Item {
				Layout.fillWidth: true
				height: bar.height
				
				ProgressBar {
					anchors.fill: parent
					id: bar
					value: 0.23
				}
				Label {
					anchors.left: parent.left
					anchors.right: parent.right
					anchors.verticalCenter: parent.verticalCenter
					anchors.margins: 4
					text: "Progress"
					elide: Text.ElideRight
				}
			}
		}

		ComicInfoPanel {}
	}
}
