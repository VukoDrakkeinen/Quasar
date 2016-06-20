import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QuasarGUI 1.0

SplitView {	
	id: root
	orientation: Qt.Horizontal
	focus: true
	property var comicId: updateModel.comicId	//Somebody explain to me why the fuck can't it be a goddamn alias
	
	Binding {
		target: updateModel
		property: "currentRow"
		value: comicListView.currentRow
	}
	
	ControlButtons {
		id: cb
		function comicIdValid() { return comicListView.currentRow != -1 }
		Button {
			text: qsTr("Add comic")
			action: Action {
				onTriggered: addComic.resetAndShow()
			}
		}
		Button {
			text: qsTr("Quick add")	//TODO: flash on first run
			enabled: false
		}
		Button {
			text: qsTr("Remove comic")
			action: Action {
				onTriggered: {
					var indices = []
					var names = []
					comicListView.selection.forEach(function (i) {
						indices.push(i)
						names.push(comicListView.model.qmlGet(i, 0, "display"))
					})
					removeComic.resetAndShow(indices, names)
				}
			}
			Component.onCompleted: this.enabled = Qt.binding(cb.comicIdValid)
		}
		Button {
			text: qsTr("Search")
			enabled: false
		}
		Button {
			text: qsTr("Select all")
			action: Action {
				onTriggered: {
					if (comicListView.currentRow == -1) {
						comicListView.currentRow = 0
					}
					comicListView.selection.selectAll()
				}
			}
		}
		Button {
			text: qsTr("Check for updates")
			action: Action {
				onTriggered: {
					//console.log(comicListView.model.data(comicListView.model.index(0, 0), 0))
					var comicIndices = []
					comicListView.selection.forEach(function (i){comicIndices.push(i)})
					quasarCore.updateComics(comicIndices)
				}
			}
			Component.onCompleted: this.enabled = Qt.binding(cb.comicIdValid)
		}
		Button {
			text: qsTr("Chapters")
			action: Action {
				onTriggered: stos.push(chapterInterface)
			}
			Component.onCompleted: this.enabled = Qt.binding(cb.comicIdValid)
		}
		Button {
			text: qsTr("Properties")
			action: Action {
				onTriggered: properties.resetAndShow()
			}
			Component.onCompleted: this.enabled = Qt.binding(cb.comicIdValid)
		}
		Button {
			text: "print value"
			action: Action {
				onTriggered: console.log(justGoInt)
			}
		}
		Button {
            text: "print object"
            action: Action {
                onTriggered: console.log(quasarCore.globalSettings.notificationMode)
            }
        }
        Button {
            text: "object value"
            action: Action {
                onTriggered: console.log(quasarCore.globalSettings.notificationMode|0)
            }
        }
		
	}

	Keys.onPressed: {
		if (event.modifiers != Qt.ControlModifier) {
			rtf.text = event.text
			rtf.focus = true
		}
	}
	
	ColumnLayout {
		Layout.fillWidth: true
		
		ComicListView {
			Layout.fillHeight: true
			Layout.fillWidth: true
			id: comicListView
			Binding {
				target: chapterModel
				property: "comicId"
				value: root.comicId
			}
		}

		TextField {
			id: rtf
			Layout.fillWidth: true
			placeholderText: qsTr("Quick search")
			onLengthChanged: queryDoneTimer.restart()
			onEditingFinished: {
				regexp.pattern = rtf.text
				queryDoneTimer.stop()
			}
			Timer {
				id: queryDoneTimer
				interval: 500;
				onTriggered: regexp.pattern = rtf.text
            }
			RegExp {
				id: regexp
				caseSensitive: false
			}
			Binding {
				target: updateModel
				property: "filterRegExp"
				when: regexp.valid
				value: regexp.regexp
			}
			Behavior on textColor {
				ColorAnimation {
					duration: 400
				}
			}
			Binding {
				target: rtf
				property: "textColor"
				when: !regexp.valid
				value: "red"
			}
		}
		
		LabeledProgressBar {
			//indeterminate: true
			enabled: false
		}
	}
	
	ComicInfoPanel {
		model: infoModel
		comicId: root.comicId
	}
} 
