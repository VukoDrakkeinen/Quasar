import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QuasarGUI 1.0

Item {
	implicitHeight: layout.height
	implicitWidth: layout.width
	
	property alias mode: disablingExclusiveGroup.currentIndex
	property alias accumulationCount: accumulationSpin.value
	property alias delayedHours: spinHours.value
	property alias delayedDays: spinDays.value
	property alias delayedWeeks: spinWeeks.value
	
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
			disablee: delaySpinBoxes
		}
		
		
		RowLayout {
			id: delaySpinBoxes
			ValuesValidator {
				id: dateValidator
			}
			ValidatingSpinBox {
				id: spinHours
				suffix: " hours"
				maximumValue: 23
				validator: dateValidator
			}
			ValidatingSpinBox {
				id: spinDays
				suffix: " days"
				maximumValue: 6
				validator: dateValidator
			}
			ValidatingSpinBox {
				id: spinWeeks
				suffix: " weeks"
				value: 1
				maximumValue: 999
				validator: dateValidator
			}
			/*
			ValidatingSpinBox {
				Layout.minimumWidth: 55
				id: spinMonths
				suffix: "m"
				maximumValue: 11
				validator: dateValidator
			}
			ValidatingSpinBox {
				Layout.minimumWidth: 60
				id: spinYears
				suffix: "Y"
				maximumValue: 290	//larger values overflow Go's time.Duration
				validator: dateValidator
			}//*/
		}
	}
}
