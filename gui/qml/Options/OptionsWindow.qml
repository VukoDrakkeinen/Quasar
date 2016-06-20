import QtQuick 2.5
import QtQuick.Window 2.2
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
//import "../utils.js" as U

Window {
	title: qsTr("Options")
	flags: Qt.Dialog
	modality: Qt.WindowModal
	color: colorOf.window
	id: root
	width: 600
	height: 500
	minimumWidth: mainLayout.Layout.minimumWidth + 2 * margin
	minimumHeight: mainLayout.Layout.minimumHeight + 2 * margin
	
	SystemPalette {
		id: colorOf
	}
	
	function __setSettings(settings) {
		notifChooser.mode = settings.notificationMode|0
		notifChooser.accumulationCount = settings.accumulativeModeCount
		var duration = settings.delayedModeDuration
		notifChooser.delayedHours = duration.hours|0
		notifChooser.delayedDays = duration.days|0
		notifChooser.delayedWeeks = duration.weeks|0
		
		downloadsPath.text = settings.downloadsPath
		fonCheckBox.checked = settings.fetchOnStartup
		intFetCheckBox.checked = settings.intervalFetching
		maxConnSpinBox.value = settings.maxConnectionsToHost
		
		var frequency = settings.fetchFrequency
		frequencySpin.hours = frequency.hours
		frequencySpin.days = frequency.days
		frequencySpin.weeks = frequency.weeks
		
		langOpts.setLanguages(settings.languages)
	}
	
	function resetAndShow() {
		this.__reset()
		this.show()
	}
	
	function __reset() {
		this.__setSettings(quasarCore.globalSettings)
	}
	
	function __defaults() {
		this.__setSettings(quasarCore.defaultGlobalSettings)
	}

	ColumnLayout {
		anchors.fill: parent
		anchors.margins: 8
		id: mainLayout

		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Default update notification mode:")
			NotificationModeChooser {
				id: notifChooser
			}
		}
		
		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Fetch settings")
			
			GridLayout {
				Layout.fillWidth: true
				Layout.fillHeight: true
				columns: 2
				
				Label {
					text: qsTr("Fetch on startup:")
					MouseArea {
						anchors.fill: parent
						onReleased: {
							fonCheckBox.pressed = false
							fonCheckBox.checked = !fonCheckBox.checked
						}
						onPressed: {
							fonCheckBox.pressed = true
						}
					}
				}
				CheckBox {
					Layout.fillWidth: true
					id: fonCheckBox
				}
				
				Label {
					text: qsTr("Interval fetching:")
					MouseArea {
						anchors.fill: parent
						onReleased: {
							intFetCheckBox.pressed = false
							intFetCheckBox.checked = !intFetCheckBox.checked
						}
						onPressed: {
							intFetCheckBox.pressed = true
						}
					}
				}
				CheckBox {
					Layout.fillWidth: true
					id: intFetCheckBox
				}
				
				Label {
					text: qsTr("Fetch frequency:")
				}
				DurationChooser {
					id: frequencySpin
					enabled: intFetCheckBox.checked
				}
				
				Label {
					text: qsTr("Max connections to host:")
				}
				SpinBox {
					id: maxConnSpinBox
					minimumValue: 1
					maximumValue: 10
				}
			}
		}

		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Downloads location:")
			RowLayout {
				anchors.fill: parent

				TextField {
					id: downloadsPath
					Layout.fillWidth: true
					text: "/home/vuko/Downloads"
				}
				Button {
					text: "IKONA"
				}
			}
		}

		GroupBox {
			Layout.fillWidth: true
			Layout.fillHeight: true
			title: qsTr("Plugins:")
			RowLayout {
				anchors.fill: parent
				Item {
					width: buttoncol.width
					height: buttoncol.height
					Layout.alignment: Qt.AlignLeft | Qt.AlignTop
					ColumnLayout {
						id: buttoncol
						Button {
							text: qsTr("Install plugin")
							Layout.fillWidth: true
						}
						Button {
							text: qsTr("Disable plugin")
							Layout.fillWidth: true
						}
						Button {
							text: qsTr("Remove plugin")
							Layout.fillWidth: true
						}
					}
				}

				TableView {
					Layout.fillHeight: true
					Layout.fillWidth: true
					Layout.minimumWidth: 300
					model: pluginsModel
					
					ListModel {
						id: pluginsModel
					}

					TableViewColumn {
						role: "name"
						title: qsTr("Plugin")
						width: 100
					}

					TableViewColumn {
						role: "status"
						title: qsTr("Status")
						width: 200
					}
				}
			}
		}
		
		GroupBox {
			Layout.fillWidth: true
			title: qsTr("Languages")
			LanguageOptions {
				anchors.fill: parent
				id: langOpts
			}
		}

		OptionsBottomButtons {
			onCancel: root.hide()
			onDefaults: root.__defaults()
			onOK: {
				var settings = {
					"fetchOnStartup": fonCheckBox.checked,
					"intervalFetching": intFetCheckBox.checked,
					"fetchFrequency": {"hours": frequencySpin.hours, "days": frequencySpin.days, "weeks": frequencySpin.weeks},
					"maxConnectionsToHost": maxConnSpinBox.value,
					"notificationMode": notifChooser.mode,
					"accumulativeModeCount": notifChooser.accumulationCount,
					"delayedModeDuration": {"hours": notifChooser.delayedHours, "days": notifChooser.delayedDays, "weeks": notifChooser.delayedWeeks},
					"downloadsPath": downloadsPath.text,
					"plugins": {"batoto": true, "bakaUpdates": false},
					"languages": langOpts.getLanguages()
				}
				
				quasarCore.setGlobalSettings(settings)
				root.hide()
			}
		}
	}
}
