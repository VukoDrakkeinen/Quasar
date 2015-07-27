import QtQuick 2.5
import QtQuick.Layouts 1.2
import QtQuick.Controls 1.4
import QuasarGUI 1.0
import "../utils.js" as U

///Rectangle {
Item {
	id: root
	Layout.minimumWidth: implicitWidth
	implicitWidth: 250
	implicitHeight: 400
	
	property int comicId: 0
	property QtObject model: null
	
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
			Layout.preferredHeight: 200
			Layout.preferredWidth: 200
			fillMode: Image.PreserveAspectFit
			source: {
				if (root.model === null) {
					return ""
				}
				var value = root.model.qmlGet(root.comicId, 0, "decoration")
				if (typeof value == "undefined") {
					return ""
				}
				return value
			}
		}
		
		Label {
			Layout.alignment: Qt.AlignHCenter | Qt.AlignTop
			text: {
				if (root.model === null) {
					return ""
				}
				var value = root.model.qmlGet(root.comicId, 0, "display")
				if (typeof value == "undefined") {
					return ""
				}
				return value
			}
			font.bold: true
			font.pointSize: 10
		}
		
		HorizontalSeparator {}
		
		ScrollView {
			id: scroll
			Layout.fillHeight: true
			Layout.fillWidth: true
			horizontalScrollBarPolicy: Qt.ScrollBarAlwaysOff
			viewport.width: 100
			GridLayout {
				width: scroll.width - 25	//too bad you can't do "width: scrollview.viewport.width - margin * 2" - sometimes the bindings start to loop :(
				columns: 2
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "AKA: "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 1, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Author(s): "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 2, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Artist(s): "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 3, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Genres: "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 4, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Categories: "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 5, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Type: "
				}
				Label {
					text: {
						if (root.model === null) {
							return ""
						}
						var status = root.model.qmlGet(root.comicId, 6, "display")
						if (status == ComicType.Invalid) { return qsTr("Invalid Comic") }
						if (status == ComicType.Manga) { return qsTr("Manga (Japanese)") }
						if (status == ComicType.Manhwa) { return qsTr("Manhwa (Korean)") }
						if (status == ComicType.Manhua) { return qsTr("Manhua (Chinese)") }
						if (status == ComicType.Western) { return qsTr("Western") }
						if (status == ComicType.Webcomic) { return qsTr("Webcomic") }
						if (status == ComicType.Other) { return qsTr("Other") }
						return "????"
					}
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Status: "
				}
				Label {
					text: {
						if (root.model === null) {
							return ""
						}
						var status = root.model.qmlGet(root.comicId, 7, "display")
						if (status == ComicStatus.Invalid) { return qsTr("Invalid Status") }
						if (status == ComicStatus.Complete) { return qsTr("Complete") }
						if (status == ComicStatus.Ongoing) { return qsTr("Ongoing") }
						if (status == ComicStatus.OnHiatus) { return qsTr("On hiatus") }
						if (status == ComicStatus.Discontinued) { return qsTr("Discontinued") }
						return "????"
					}
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Scanlation: "
				}
				Label {
					text: {
						if (root.model === null) {
							return ""
						}
						var status = root.model.qmlGet(root.comicId, 6, "display")
						if (status == ScanlationStatus.Invalid) { return qsTr("Invalid Scanlation Status") }
						if (status == ScanlationStatus.Complete) { return qsTr("Complete") }
						if (status == ScanlationStatus.Ongoing) { return qsTr("Ongoing") }
						if (status == ScanlationStatus.OnHiatus) { return qsTr("On hiatus") }
						if (status == ScanlationStatus.Dropped) { return qsTr("Dropped") }
						if (status == ScanlationStatus.InDesperateNeedOfMoreStaff) { return qsTr("IN DESPERATE NEED OF MORE STAFF") }
						return "????"
					}
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Mature: "
				}
				Label {
					text: {
						if (root.model === null) {
							return ""
						}
						root.model.qmlGet(root.comicId, 9, "display") ? "Yes" : "No"
					}
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Rating: "
				}
				Label {
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 10, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
				}
				
				Label {
					Layout.alignment: Qt.AlignRight | Qt.AlignTop
					text: "Description: "
				}
				Label {
					Layout.fillWidth: true
					text: {
						if (root.model === null) {
							return ""
						}
						var value = root.model.qmlGet(root.comicId, 11, "display")
						if (typeof value == "undefined") {
							return ""
						}
						return value
					}
					wrapMode: Text.Wrap
				}
			}
		}
	}
	
	states: [
		State {
			name: "hidden"
			PropertyChanges { target: root; width: 0 }
			PropertyChanges { target: root; Layout.maximumWidth: 0 }
			PropertyChanges { target: root; Layout.minimumWidth: 0 }
		}
	]
	state: "hidden"
	
	transitions: Transition {
		NumberAnimation { properties: "width,Layout.maximumWidth,Layout.minimumWidth"; duration: 200 }
	}
	
	onComicIdChanged: {
		if (this.comicId != -1) {
			this.state = ""
		} else {
			this.state = "hidden"
		}
	}
}
