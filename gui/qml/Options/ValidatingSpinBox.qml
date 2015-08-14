import QtQuick 2.5
import QtQuick.Controls 1.4
import QtQml 2.2
import QuasarGUI 1.0

SpinBox {
	property var validator: null
	QtObject {
		id: internal
		property var prevValidator: null
	}
	
	onValidatorChanged: {
		if (internal.prevValidator) {
			validator.unbindObject(this)
		}
		internal.prevValidator = validator
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