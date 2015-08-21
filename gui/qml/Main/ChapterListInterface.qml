import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QtQml.Models 2.2

SplitView {
	orientation: Qt.Horizontal
	
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
				onTriggered: stos.push(readInterface)
			}
		}
		Button {
			text: qsTr("Read selected")
			action: Action {
				onTriggered: stos.push(readInterface)
			}
		}
		Button {
			text: qsTr("Read last")
			action: Action {
				onTriggered: stos.push(readInterface)
			}
		}
		Button {
			text: qsTr("Download selected")
			action: Action {
				onTriggered: console.log(quasarCore.pluginNames())
			}
		}
		Button {
			text: qsTr("Select all")
			action: Action {
				onTriggered: chapterListView.selection.select(chapterModel.index(0, 0), ItemSelectionModel.ClearAndSelect | ItemSelectionModel.Rows | ItemSelectionModel.Columns)
			}
		}
		Button {
			text: qsTr("Mark as read")
			action: Action {
				onTriggered: console.log(chapterListView.selection.selectedIndexes())	//Need a conversion
				//onTriggered: chapterListView.selection.forEach(function (i){console.log("Chapter", i, "marked as read")})
			}
		}
		Button {
			text: qsTr("Mark as unread")
			action: Action {
				onTriggered: console.log(chapterListView.selection.selectedIndexes())
				//onTriggered: chapterListView.selection.forEach(function (i){console.log("Chapter", i, "marked as unread")})
			}
		}
		
	}
	
	ChapterListView { id: chapterListView }
}

