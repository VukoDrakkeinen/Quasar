import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtQuick.Controls 1.4

Rectangle {
	id: root
	width: 600
	height: 400
	color: colorOf.window
	
	property int comicId: -1
	property int chapterId: 0
	property int scanlationId: 0
	
	SystemPalette { id: colorOf }
	
	RowLayout {
		anchors.fill: parent
		
		ScrollView {
			//horizontalScrollBarPolicy: Qt.AlwaysOn
			Layout.fillHeight: true
			Column {
				
				Image {
					source: "placeholder.png"
					Rectangle {
						color: "black"
						anchors.left: parent.left
						anchors.top: parent.top
						anchors.margins: 8
						width: ll.width
						height: ll.height
						Label {
							id: ll
							color: "white"
							text: "01"
							font.pointSize: 8
						}
					}
				}
				Image {
					source: "placeholder.png"
					Rectangle {
						color: "black"
						anchors.left: parent.left
						anchors.top: parent.top
						anchors.margins: 8
						width: lla.width
						height: lla.height
						Label {
							id: lla
							color: "white"
							text: "02"
							font.pointSize: 8
						}
					}
				}
				Image {
					source: "placeholder.png"
				}
			}
		}
		
		Item {
			Text {
				anchors.centerIn: parent
				font.pointSize: 30
				color: Qt.rgba(1, 0, 0, 1)
				text: "Comic " + root.comicId + ", chapter " + root.chapterId + "[" + root.scanlationId + "]"
			}
			Layout.fillWidth: true
			Layout.fillHeight: true
			/*Image {
				anchors.centerIn: parent
				scale: 5
				source: "placeholder.png"
			}*/
		}
	}
}