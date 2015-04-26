import QtQuick 2.0
import QtQuick.Controls 1.1
import QtQuick.Layouts 1.1
import QuasarGUI 1.0

Item {
	implicitHeight: layout.height
	implicitWidth: layout.width
	
	//objectName: "czuzer"
	
	signal componentCompleted()
	Component.onCompleted: componentCompleted
	
	function setValues(mode, count, duration) {
		notifmode.data[1].buttons[mode].checked = true
		accumulationSpin.value = count
		//TODO: datetime
	}
	
	
	GridLayout {
		id: layout
		columns: 2
		
		DisablingExclusiveGroup {
			id: notifmode
		}
		DisablingRadioButton {
			Layout.columnSpan: 2
			text: qsTr("Immediate")
			disabler: notifmode
			checked: true
		}
		
		DisablingRadioButton {
			text: qsTr("Accumulative:")
			disabler: notifmode
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
			disabler: notifmode
			disablee: delaySpinBoxes
		}
		
		
		RowLayout {
			id: delaySpinBoxes
			ValuesValidator {
				id: dateValidator
			}
			ValidatingSpinBox {
				Layout.minimumWidth: 55	//FIXME: hardcoded values
				id: spinHours
				suffix: "H"
				maximumValue: 23
				validator: dateValidator
			}
			ValidatingSpinBox {
				Layout.minimumWidth: 45
				id: spinDays
				suffix: "d"
				maximumValue: 6
				validator: dateValidator
			}
			ValidatingSpinBox {
				Layout.minimumWidth: 45
				id: spinWeeks
				suffix: "w"
				value: 1
				maximumValue: 3
				validator: dateValidator
			}
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
			}
		}
		
		DisablingRadioButton {
			Layout.columnSpan: 2
			text: qsTr("Manual")
			disabler: notifmode
		}
	}
}
