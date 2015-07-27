import QtQuick 2.5
import QtQuick.Controls 1.4

RadioButton {
    property DisablingExclusiveGroup disabler: null
    property var disablee: null

    QtObject {
		id: internal
		property DisablingExclusiveGroup prevDisabler: null
    }

    onDisablerChanged: {
        if (disabler) {
			if (disabler !== internal.prevDisabler) {
				if (internal.prevDisabler) {
					internal.prevDisabler.unbindCheckable(this)
					exclusiveGroup.unbindCheckable(this)
				}
				internal.prevDisabler = disabler
				this.exclusiveGroup = disabler.__internalGroup
				disabler.bindCheckable(this)
			}
        }
    }
    
    onDisableeChanged: {
		__toggleDisablee()
    }

    onCheckedChanged: {
		__toggleDisablee()
    }
    
    function __toggleDisablee() {
		if (disablee) {
			disablee.enabled = this.checked
		}
    }
}
