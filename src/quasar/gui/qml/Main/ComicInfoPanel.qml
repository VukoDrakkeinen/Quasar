import QtQuick 2.0
import QtQuick.Layouts 1.1
import QtQuick.Controls 1.1

///Rectangle {
Item {
	Layout.minimumWidth: implicitWidth
	implicitWidth: 200
	implicitHeight: 600
	
	///color: colorOf.window
	///SystemPalette { id: colorOf }
	
	ColumnLayout {
		id: layout
		anchors.fill: parent
		anchors.margins: 5
		
		Label {
			Layout.alignment: Qt.AlignLeft | Qt.AlignTop
			text: "Information"
		}
		
		Image {
			Layout.alignment: Qt.AlignHCenter | Qt.AlignVCenter
			source: "file:///home/vuko/Pictures/Misc/Azureus.png"
		}
		
		Label {
			Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
			text: "HorizontalSeparator.qml"
			font.bold: true
			font.pointSize: 10
		}
		
		HorizontalSeparator {}
		
		GridLayout {
			Layout.fillWidth: true
			Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
			columns: 2
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "AKA: "
			}
			Label {
				Layout.fillWidth: true
				text: "Also Known As Many Other Titles"
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Author(s): "
			}
			Label {
				Layout.fillWidth: true
				text: "Authoring Author"
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Artist(s): "
			}
			Label {
				Layout.fillWidth: true
				text: "Artisting Artist"
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Genres: "
			}
			Label {
				Layout.fillWidth: true
				text: "[Action], [Mystery], [Drama]"
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Categories: "
			}
			Label {
				Layout.fillWidth: true
				text: "[Seinen]"
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Type: "
			}
			Label {
				text: "Manga"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Status: "
			}
			Label {
				text: "Complete"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Scanlation: "
			}
			Label {
				text: "Ongoing"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Mature: "
			}
			Label {
				text: "Yes"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Rating: "
			}
			Label {
				text: "9.2"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Description: "
			}
			Label {
				Layout.fillWidth: true
				text: "A very long description of what is essentially lorem ipsum shit"
				wrapMode: Text.Wrap
			}
		}
	}
}
