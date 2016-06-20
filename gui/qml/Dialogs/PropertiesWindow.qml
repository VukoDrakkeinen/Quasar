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
		notifChooser.mode = settings.notificationMode|0
		notifChooser.accumulationCount = settings.accumulativeModeCount|0
		var duration = settings.delayedModeDuration
		notifChooser.delayedHours = duration.hours|0
		notifChooser.delayedDays = duration.days|0
		notifChooser.delayedWeeks = duration.weeks|0
    }
	
	function resetAndShow() {
		sourcesModel.clear()
		var names = quasarCore.pluginNames()
		sourcesView.populatePluginNames(Array.prototype.concat("[autodetect]", names[0]))
		sourcesView.populatePluginComboBox(Array.prototype.concat(qsTr("[Autodetect]"), names[1]))
		sourcesView.clearFields()
		
		var sources = quasarCore.comicSources(this.comicId)
		var mappedSources = sources.map(function (source) {
			return {"priority": -1, "sourceIdx": sourcesView.indexForPluginName(source.sourceId), "url": source.url, "markAsRead": source.markAsRead}
		})
		sourcesModel.append(mappedSources)
		
		var cfg = quasarCore.comicConfig(this.comicId)
		this.__setSettings(cfg)
		this.show()
	}
	
	function __defaults() {
		var settings = quasarCore.globalSettings
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
		
		//TODO: scanlators priority

        OptionsBottomButtons {
			onCancel: root.hide()
			onDefaults: root.__defaults()
			onOK: {
				var cfg = {
					"notificationMode": notifChooser.mode,
					"accumulativeModeCount": notifChooser.accumulationCount,
					"delayedModeDuration": {"hours": notifChooser.delayedHours, "days": notifChooser.delayedDays, "weeks": notifChooser.delayedWeeks}
				}
				var sources = []
				for (var i = 0; i < sourcesModel.count; i++) {
					var item = sourcesModel.get(i);
					sources.push({"sourceId": sourcesView.pluginNameForIndex(item.sourceIdx), "url": item.url, "markAsRead": item.markAsRead})
				}
				
				quasarCore.setComicConfigAndSources(root.comicId, cfg, sources)
				root.hide()
			}
		}
    }
}
