import QtQuick 2.1
import QtQuick.Controls 1.1
import QtQuick.Controls.Styles 1.1
import QtQuick.Window 2.0
import QtQuick.Layouts 1.1
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
			TableView {
				Layout.fillHeight: true
				Layout.fillWidth: true
				implicitWidth: 500
				selectionMode: 2 //SelectionMode.ExtendedSelection
				model: comicListModel
				itemDelegate: Label{
					anchors.fill:parent
					text: model.qmlGet(styleData.row, styleData.column, "display")
					color: model.qmlGet(styleData.row, styleData.column, "foreground")
					elide: styleData.elideMode
				}

				TableViewColumn {
					title: "Title"
					width: 200
				}

				TableViewColumn {
					title: "Chapters"
					/*delegate: Label{
							text: model.qmlGet(styleData.row, styleData.column, "display")
							color: model.qmlGet(styleData.row, styleData.column, "foreground")
						}*/
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
					/*delegate: ProgressBar{
						anchors.fill: parent
						maximumValue: 100
						value: model.qmlGet(styleData.row, styleData.column, "progress")
					}//*/
					delegate: Item {
											anchors.fill: parent

											ProgressBar {
												anchors.fill: parent
												maximumValue: 100
												/*style: ProgressBarStyle {
															background: Rectangle {
																radius: 2
																color: "transparent"
																implicitWidth: 200
																implicitHeight: 24
															}
														}//*/
												value: {
													var status = model.qmlGet(styleData.row, styleData.column, "status")
													if (status == 0) return 0
													if (status == 2) return 0
													if (status == 3) return 0
													return model.qmlGet(styleData.row, styleData.column, "progress")
												}
											}
											Label {
												anchors.left: parent.left
												anchors.right: parent.right
												anchors.verticalCenter: parent.verticalCenter
												anchors.margins: 4
												text: {
													var status = model.qmlGet(styleData.row, styleData.column, "status")
													if (status == 0) return "No Updates"
													if (status == 1) return "Updating..."
													if (status == 2) return "New Chapters"
													if (status == 3) return "ERROR"
													return "???"
												}
												color: model.qmlGet(styleData.row, styleData.column, "foreground")
												elide: Text.ElideRight
											}
										}
					title: "Status"
					width: 200
				}
			}

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
