import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQml 2.2
import QuasarGUI 1.0

SpinBox {
	property ValuesValidator validator: null
	property ValuesValidator __prev_validator: null
	
	onValidatorChanged: {
		if (__prev_validator) {
			validator.unbindObject(this)
		}
		__prev_validator = validator
		if (validator) {
			validator.bindObject(this)
		}
	}
	
	Connections {
		onValueChanged: {
			if (validator) {
				validator.work()
			}
		}
	}
}