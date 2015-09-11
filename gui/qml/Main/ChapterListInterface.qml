import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QtQml.Models 2.2
import QtQml 2.2
import QuasarGUI 1.0

SplitView {
	id: root
	orientation: Qt.Horizontal
	
	property int comicId: -1
	
	QtObject {
		id: internal
		property var nullIndex: null
		Component.onCompleted: internal.nullIndex = chapterListView.model.index(-1, -1)
	}
	
	ControlButtons {
		Button {
			text: qsTr("Back")
			action: Action {
				onTriggered: stos.pop()
			}
		}
		Button {
			text: qsTr("Read next")
			action: Action {
				onTriggered: {
					var chapterIdx = quasarCore.getQueuedChapter(root.comicId)
					readInterface.chapterId = chapterIdx
					readInterface.scanlationId = 0
					stos.push(readInterface)
				}
			}
		}
		Button {
			id: readSelBtn
			text: qsTr("Read selected")
			action: Action {
				onTriggered: {
					var idx = ModelListConverter.convert(chapterListView.selection.currentIndex)
					readInterface.chapterId = idx.x
					readInterface.scanlationId = idx.y
					stos.push(readInterface)
				}
			}
			Binding on enabled {	//WORKAROUND
				value: (root.comicId != root.comicId) || chapterListView.selection.currentIndex != internal.nullIndex	//trigger on comicId change
			}
		}
		Button {
			text: qsTr("Read last")
			action: Action {
				onTriggered: {
					var chapterIdx = quasarCore.getLastReadChapter(root.comicId)
					readInterface.chapterId = chapterIdx
					readInterface.scanlationId = 0
					stos.push(readInterface)
				}
			}
		}
		Button {
			text: qsTr("Download selected")
			action: Action {
				onTriggered: {
					var conv = ModelListConverter.convertMany(chapterListView.selection.selectedIndexes())
					var chapterIndices = conv.map(function(item) {
						return item.x
					})
					var scanlationIndices = conv.map(function(item) {
						return item.y
					})
					quasarCore.downloadPages(root.comicId, chapterIndices, scanlationIndices)
					console.log("Download", conv, "of comic", root.comicId)
				}
			}
			Binding on enabled {	//WORKAROUND
				value: (root.comicId != root.comicId) || chapterListView.selection.currentIndex != internal.nullIndex	//trigger on comicId change
			}
		}
		Button {
			text: qsTr("Select all")
			action: Action {
				onTriggered: {
					chapterListView.selection.setCurrentIndex(chapterModel.index(0, 0), ItemSelectionModel.Select)
					chapterListView.selection.select(chapterModel.index(0, 0), ItemSelectionModel.ClearAndSelect | ItemSelectionModel.Columns)
				}
			}
		}
		Button {
			text: qsTr("Mark as read")
			action: Action {
				onTriggered: {
					var list = ModelListConverter.convertMany(chapterListView.selection.selectedIndexes()).map(function(item) {
						return item.x
					})
					quasarCore.markAsRead(root.comicId, list, true)
				}
			}
			Binding on enabled {	//WORKAROUND
				value: (root.comicId != root.comicId) || chapterListView.selection.currentIndex != internal.nullIndex	//trigger on comicId change
			}
		}
		Button {
			text: qsTr("Mark as unread")
			action: Action {
				onTriggered: {
					var list = ModelListConverter.convertMany(chapterListView.selection.selectedIndexes()).map(function(item) {
						return item.x
					})
					quasarCore.markAsRead(root.comicId, list, false)
				}
			}
			Binding on enabled {	//WORKAROUND
				value: (root.comicId != root.comicId) || chapterListView.selection.currentIndex != internal.nullIndex	//trigger on comicId change
			}
		}
		
	}
	
	ChapterListView { id: chapterListView }
	data: [
		ReadingInterface {
			id: readInterface
			visible: false
			comicId: root.comicId
		}
	]
}

