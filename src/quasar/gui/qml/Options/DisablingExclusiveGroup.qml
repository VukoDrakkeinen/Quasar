import QtQuick 2.5
import QtQuick.Controls 1.4

Item {
	ExclusiveGroup {
		id: internalGroup
	}
	Item {
		id: container
		property var buttons: []
	}
	
	function bindButton(object) {
		if (!container.buttons)
			container.buttons = []
			
			container.buttons.unshift(object)
			disableUnchecked()
	}
	
	function unbindButton(object) {
		if (!container.buttons) {
			container.buttons = []
		}
		
		var i = container.buttons.indexOf(object)
		container.buttons.splice(i, i + 1)
	}
	
	function disableUnchecked() {
		if (!container.buttons) {
			container.buttons = []
		}
		for (var i = 0; i < container.buttons.length; i++) {
			var button = container.buttons[i]
			var disablee = button.disablee
			if (disablee) {
				button.disablee.enabled = button.checked
			}
		}
	}
}
