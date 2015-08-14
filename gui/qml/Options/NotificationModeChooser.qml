import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QuasarGUI 1.0

Item {
	implicitHeight: layout.height
	implicitWidth: layout.width
	
	property alias mode: disablingExclusiveGroup.currentIndex
	property alias accumulationCount: accumulationSpin.value
	property alias delayedHours: delaySpin.hours
	property alias delayedDays: delaySpin.days
	property alias delayedWeeks: delaySpin.weeks
	
	GridLayout {
		id: layout
		columns: 2
		
		DisablingExclusiveGroup {
			id: disablingExclusiveGroup
		}
		DisablingRadioButton {
			Layout.columnSpan: 2
			text: qsTr("Immediate")
			disabler: disablingExclusiveGroup
			checked: true
		}
		
		DisablingRadioButton {
			text: qsTr("Accumulative:")
			disabler: disablingExclusiveGroup
			disablee: accumulationSpin
		}
		SpinBox {
			id: accumulationSpin
			suffix: qsTr(" chapters")
			minimumValue: 2
			Layout.minimumWidth: 110
		}
		
		DisablingRadioButton {
			text: qsTr("Delayed:")
			disabler: disablingExclusiveGroup
			disablee: delaySpin
		}
		
		
		DurationChooser {
			id: delaySpin
		}
	}
}
