import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2

SplitView {
	orientation: Qt.Horizontal
	
	property alias comicId: view.comicId
	property alias chapterId: view.chapterId
	property alias scanlationId: view.scanlationId
	
	ControlButtons {
		Button {
			text: qsTr("Back")
			action: Action {
				onTriggered: stos.pop()
			}
		}
	}
	
	ColumnLayout {
		Layout.fillWidth: true
		
		ComicView {
			id: view
		}
	}
} 

