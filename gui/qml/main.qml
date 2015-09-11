import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Window 2.2
import QtQuick.Layouts 1.2
import "Main"
import "Options"
import "Dialogs"
import "utils.js" as U

ApplicationWindow {
	title: "Quasar"
	property int margin: 11
	width: 1200 + 2 * margin
	height: 800 + 2 * margin
	minimumWidth: 500 + 2 * margin
	minimumHeight: 500 + 2 * margin

	OptionsWindow {
		id: options
	}

	PropertiesWindow {
		id: properties
		comicId: comicInterface.comicId
	}
	
	AddComicWindow {
		id: addComic
	}
	
	RemoveComicDialog {
		id: removeComic
	}

	menuBar: MenuBar {
		Menu {
			title: "Quasar"
			MenuItem {
				text: qsTr("Options")
				onTriggered: options.resetAndShow()
			}
			MenuItem {
				text: qsTr("Quit")
				onTriggered: Qt.quit()
			}
		}
	}
	
	StackView {
		id: stos
		anchors.fill: parent
		ChapterListInterface{ id: chapterInterface; visible: false; comicId: comicInterface.comicId  }
		ComicListInterface{ id: comicInterface; visible: false }
		initialItem: comicInterface
	}
}
