import QtQuick 2.5
import QtQuick.Controls 1.4

Item {
	id: root
	
	ExclusiveGroup {
		id: internalGroup
		onCurrentChanged: root.__updateCurrentIndex()
	}
	
	QtObject {
		id: internal
		property var buttons: []
		property bool breakRecurrence: false
	}
	
	property int currentIndex: -1
	property alias __internalGroup: internalGroup
	
	function bindCheckable(object) {
		internal.buttons.unshift(object)
		this.__updateCurrentIndex()
	}
	
	function unbindCheckable(object) {
		if (object.disablee) {
			object.disablee.enabled = true
		}
		
		var i = internal.buttons.indexOf(object)
		internal.buttons.splice(i, i + 1)
	}
	
	function __updateCurrentIndex() {
		internal.breakRecurrence = true
		for (var i = 0; i < internal.buttons.length; i++) {
			if (internal.buttons[i].checked) {
				this.currentIndex = i
				break
			}
		}
	}
	
	onCurrentIndexChanged: {
		if (internal.breakRecurrence) {
			internal.breakRecurrence = false
			return
		}
		
		if (currentIndex >= internal.buttons.length) {
			console.warn("DisablingExclusiveGroup: currentIndex out of bounds")
			currentIndex = -1
			return
		}
		internal.buttons[currentIndex].checked = true
	}
}
