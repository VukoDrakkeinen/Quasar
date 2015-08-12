import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Options"
import "../Workarounds"

Window {
	title: qsTr("Add Comic")
	flags: Qt.Dialog
	modality: Qt.WindowModal
	color: colorOf.window
	id: thisWindow
	width: mainLayout.implicitWidth + 2 * margin + 300
	height: mainLayout.implicitHeight + 2 * margin
	minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
	minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin
	
	SystemPalette {
		id: colorOf
	}
	
	function resetAndShow() {
		sourcesModel.clear()
		var names = quasarCore.pluginNames()
		sourcesView.populatePluginNames(Array.prototype.concat("[autodetect]", names[0]))
		sourcesView.populatePluginComboBox(Array.prototype.concat(qsTr("[Autodetect]"), names[1]))
		__reset()
		this.show()
	}
	
	function __reset() {
		var settings = quasarCore.globalSettings()
		notifChooser.mode = settings.notificationMode
		notifChooser.accumulationCount = settings.accumulativeModeCount
		var duration = settings.delayedModeDuration
		notifChooser.delayedHours = duration.hours
		notifChooser.delayedDays = duration.days
		notifChooser.delayedWeeks = duration.weeks
	}
	
	ColumnLayout {
		id: mainLayout
		anchors.fill: parent
		anchors.margins: 8
		
		GroupBox {
			Layout.fillWidth: true
			Layout.fillHeight: true
			title: qsTr("Sources:")
			SourcesListView {
				anchors.fill: parent
				id: sourcesView
				model: sourcesModel
				
				ListModel {
					id: sourcesModel
				}
			}
		}
		
		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Update notification mode:")
			NotificationModeChooser {
				id: notifChooser
			}
		}
		
		OptionsBottomButtons {
			onCancel: thisWindow.hide()
			onDefaults: thisWindow.__reset()
			onOK: {
				var settings = {
					"notificationMode": notifChooser.mode, "accumulativeModeCount": notifChooser.accumulationCount,
					"delayedModeDuration": {"hours": notifChooser.delayedHours, "days": notifChooser.delayedDays, "weeks": notifChooser.delayedWeeks}
				}
				
				var sources = []
				for (var i = 0; i < sourcesModel.count; i++) {
					var item = sourcesModel.get(i);
					sources.push({"pluginName": sourcesView.pluginNameForIndex(item.sourceIdx), "url": item.url, "markAsRead": item.markAsRead})
				}
				
				quasarCore.addComic(settings, sources)
				thisWindow.hide()
			}
		}
	}
}
