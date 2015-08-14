import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQuick.Layouts 1.2
import QuasarGUI 1.0

RowLayout {
	property alias hours: spinHours.value
	property alias days: spinDays.value
	property alias weeks: spinWeeks.value
	DurationValidator {
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
}
