import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2

SplitView {
	orientation: Qt.Horizontal
	
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
		
		ComicView {}
	}
} 

