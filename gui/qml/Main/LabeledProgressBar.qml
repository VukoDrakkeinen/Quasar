import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
//import QuasarGUI 1.0

Item {
	Layout.fillWidth: true
	height: bar.height
	
	property alias text: label.text
	property alias value: bar.value
	//property alias indeterminate: bar.indeterminate
	
	//SaneProgressBar {
	ProgressBar {
		anchors.fill: parent
		id: bar
	}
	Label {
		id: label
		anchors.left: parent.left
		anchors.right: parent.right
		anchors.verticalCenter: parent.verticalCenter
		anchors.margins: 4
		text: "Progress"
		elide: Text.ElideRight
	}
} 
