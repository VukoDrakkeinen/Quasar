import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtQuick.Controls 1.4

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
			//source: "/home/vuko/Pictures/Misc/Azureus.png"
			source: infoModel.qmlGet(0, 0, "decoration")
		}
		
		Label {
			Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
			//text: "HorizontalSeparator.qml"
			text: infoModel.qmlGet(0, 0, "display")
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
				//text: "Also Known As Many Other Titles"
				text: infoModel.qmlGet(0, 1, "display")
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Author(s): "
			}
			Label {
				Layout.fillWidth: true
				//text: "Authoring Author"
				text: infoModel.qmlGet(0, 2, "display")
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Artist(s): "
			}
			Label {
				Layout.fillWidth: true
				//text: "Artisting Artist"
				text: infoModel.qmlGet(0, 3, "display")
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Genres: "
			}
			Label {
				Layout.fillWidth: true
				//text: "[Action], [Mystery], [Drama]"
				text: infoModel.qmlGet(0, 4, "display")
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Categories: "
			}
			Label {
				Layout.fillWidth: true
				//text: "[Seinen]"
				text: infoModel.qmlGet(0, 5, "display")
				wrapMode: Text.Wrap
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Type: "
			}
			Label {
				//text: "Manga"
				text: {
					var status = infoModel.qmlGet(0, 6, "display")
					if (status == 0) return "Invalid Comic"
					if (status == 1) return "Manga (Japanese)"
					if (status == 2) return "Manhwa (Korean)"
					if (status == 3) return "Manhua (Chinese)"
					if (status == 4) return "Western"
					if (status == 5) return "Webcomic"
					if (status == 6) return "Other"
					return "????"
				}
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Status: "
			}
			Label {
				//text: "Complete"
				text: {
					var status = infoModel.qmlGet(0, 7, "display")
					if (status == 0) return "Invalid Status"
					if (status == 1) return "Complete"
					if (status == 2) return "Ongoing"
					if (status == 3) return "On hiatus"
					if (status == 4) return "Discontinued"
					return "????"
				}
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Scanlation: "
			}
			Label {
				//text: "Ongoing"
				text: {
					var status = infoModel.qmlGet(0, 6, "display")
					if (status == 0) return "Invalid Scanlation Status"
					if (status == 1) return "Complete"
					if (status == 2) return "Ongoing"
					if (status == 3) return "On hiatus"
					if (status == 4) return "Dropped"
					if (status == 5) return "IN DESPERATE NEED OF MORE STAFF"
					return "????"
				}
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Mature: "
			}
			Label {
				//text: "Yes"
				text: infoModel.qmlGet(0, 9, "display") ? "Yes" : "No"
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Rating: "
			}
			Label {
				//text: "9.2"
				text: infoModel.qmlGet(0, 10, "display")
			}
			
			Label {
				Layout.alignment: Qt.AlignRight | Qt.AlignTop
				text: "Description: "
			}
			Label {
				Layout.fillWidth: true
				//text: "A very long description of what is essentially lorem ipsum shit"
				text: infoModel.qmlGet(0, 11, "display")
				wrapMode: Text.Wrap
			}
		}
	}
}
