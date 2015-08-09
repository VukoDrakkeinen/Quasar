import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2

SplitView {
	orientation: Qt.Horizontal
	
	ControlButtons {
		Button {
			text: qsTr("Back")
		}
	}
	
	ColumnLayout {
		Layout.fillWidth: true
		
		Label {
			text: "Placeholder"
		}
	}
	
	ComicInfoPanel {}
} 

