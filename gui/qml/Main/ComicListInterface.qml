import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../utils.js" as U

SplitView {	
	id: root
	orientation: Qt.Horizontal
	property alias comicId: comicListView.currentRow
	
	ControlButtons {
		Button {
			text: qsTr("Add comic")
			action: Action {
				onTriggered: addComic.resetAndShow()
			}
		}
		Button {
			text: qsTr("Remove comic")
			action: Action {
				onTriggered: {
					var indices = []
					var names = []
					comicListView.selection.forEach(function (i) {
						indices.push(i)
						names.push(comicListView.model.qmlGet(i, 0, "display"))
					})
					removeComic.resetAndShow(indices, names)
				}
			}
			Component.onCompleted: this.enabled = Qt.binding(function() { return comicListView.currentRow != -1})
		}
		Button {
			text: qsTr("Select all")
			action: Action {
				onTriggered: {
					if (comicListView.currentRow == -1) {
						comicListView.currentRow = 0
					}
					comicListView.selection.selectAll()
				}
			}
		}
		Button {
			text: qsTr("Check for updates")
			action: Action {
				onTriggered: {
					console.log(comicListView.model.data(comicListView.model.index(0, 0), comicListView.model.roleNames()["foreground"]))
					comicListView.selection.forEach(function (i){console.log("Comic", i, "requested to update")})
				}
			}
			Component.onCompleted: this.enabled = Qt.binding(function() { return comicListView.currentRow != -1})
		}
		Button {
			text: qsTr("Chapters")
			action: Action {
				onTriggered: stos.push(chapterInterface)
			}
			Component.onCompleted: this.enabled = Qt.binding(function() { return comicListView.currentRow != -1})
		}
		Button {
			text: qsTr("Properties")
			action: Action {
				onTriggered: properties.resetAndShow()
			}
			Component.onCompleted: this.enabled = Qt.binding(function() { return comicListView.currentRow != -1})
		}
		
	}
	
	ColumnLayout {
		Layout.fillWidth: true
		
		ComicListView {
			Layout.fillHeight: true
			Layout.fillWidth: true
			id: comicListView
			onCurrentRowChanged: chapterModel.setComicIdx(this.currentRow)
		}
		
		LabeledProgressBar {}
	}
	
	ComicInfoPanel {
		model: infoModel
		comicId: root.comicId
	}
} 
