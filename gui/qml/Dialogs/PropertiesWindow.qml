import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import "../Options"
import "../Workarounds"

Window {
    title: qsTr("Comic Properties")
    flags: Qt.Dialog
    modality: Qt.WindowModal
    color: colorOf.window
    id: root
    width: mainLayout.implicitWidth + 2 * margin + 300
    height: mainLayout.implicitHeight + 2 * margin
    minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
    minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin
    
    property int comicId: 0

    SystemPalette {
        id: colorOf
    }
    
    function __setSettings(settings) {
		notifChooser.mode = settings.notificationMode
		notifChooser.accumulationCount = settings.accumulativeModeCount
		var duration = settings.delayedModeDuration
		notifChooser.delayedHours = duration.hours
		notifChooser.delayedDays = duration.days
		notifChooser.delayedWeeks = duration.weeks
    }
	
	function resetAndShow() {
		sourcesModel.clear()
		var names = quasarCore.pluginNames()
		sourcesView.populatePluginNames(Array.prototype.concat("[autodetect]", names[0]))
		sourcesView.populatePluginComboBox(Array.prototype.concat(qsTr("[Autodetect]"), names[1]))
		sourcesView.clearFields()
		
		var sources = quasarCore.comicSources(this.comicId)
		var mappedSources = sources.map(function (source) {
			return {"priority": -1, "sourceIdx": sourcesView.indexForPluginName(source.pluginName), "url": source.url, "markAsRead": source.markAsRead}
		})
		sourcesModel.append(mappedSources)
		
		var settings = quasarCore.comicSettings(this.comicId)
		this.__setSettings(settings)
		this.show()
	}
	
	function __defaults() {
		var settings = quasarCore.globalSettings()
		this.__setSettings(settings)
	}

    ColumnLayout {
        id: mainLayout
        anchors.fill: parent
        anchors.margins: 8

        GroupBox {
			Layout.fillWidth: true
			title: qsTr("Update notification mode:")
			NotificationModeChooser {
				id: notifChooser
			}
		}

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

        OptionsBottomButtons {
			onCancel: root.hide()
			onDefaults: root.__defaults()
			onOK: {
				var settings = {"notificationMode": notifChooser.mode, "accumulativeModeCount": notifChooser.accumulationCount}
				var sources = []
				for (var i = 0; i < sourcesModel.count; i++) {
					var item = sourcesModel.get(i);
					sources.push({"pluginName": sourcesView.pluginNameForIndex(item.sourceIdx), "url": item.url, "markAsRead": item.markAsRead})
				}
				var delayedModeDuration = {"hours": notifChooser.delayedHours, "days": notifChooser.delayedDays, "weeks": notifChooser.delayedWeeks}
				
				quasarCore.setComicSettingsAndSources(root.comicId, settings, delayedModeDuration, sources)
				root.hide()
			}
		}
    }
}
